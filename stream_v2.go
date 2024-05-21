package openai

import (
	"encoding/json"
	"io"
)

type StreamRawEvent struct {
	streamEvent
	Data json.RawMessage
}

type StreamDone struct {
	streamEvent
}

type StreamThreadMessageCompleted struct {
	Message
	streamEvent
}

type StreamThreadMessageDelta struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Delta  Delta  `json:"delta"`

	streamEvent
}

type Delta struct {
	// DeltaText | DeltaImageFile
	Content []DeltaContent `json:"content"`
}

type DeltaContent struct {
	Index int    `json:"index"`
	Type  string `json:"type"`

	Text      *DeltaText      `json:"text"`
	ImageFile *DeltaImageFile `json:"image_file"`
	ImageURL  *DeltaImageURL  `json:"image_url"`
}

type DeltaText struct {
	Value string `json:"value"`
	// Annotations []any  `json:"annotations"`
}

type DeltaImageFile struct {
	FileID string `json:"file_id"`
	Detail string `json:"detail"`
}

type DeltaImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail"`
}

func NewStreamerV2(r io.Reader) *StreamerV2 {
	var rc io.ReadCloser

	if closer, ok := r.(io.ReadCloser); ok {
		rc = closer
	} else {
		rc = io.NopCloser(r)
	}

	return &StreamerV2{
		readCloser: rc,
		scanner:    NewSSEScanner(r, false),
	}
}

type StreamerV2 struct {
	// readCloser is only used for closing the stream
	readCloser io.ReadCloser

	scanner *SSEScanner
	next    StreamEvent

	// buffer for implementing io.Reader
	buffer []byte
}

// TeeSSE tees the stream data with a io.TeeReader
func (s *StreamerV2) TeeSSE(w io.Writer) {
	// readCloser is a helper struct that implements io.ReadCloser by combining an io.Reader and an io.Closer
	type readCloser struct {
		io.Reader
		io.Closer
	}

	s.readCloser = &readCloser{
		Reader: io.TeeReader(s.readCloser, w),
		Closer: s.readCloser,
	}

	s.scanner = NewSSEScanner(s.readCloser, false)
}

// Close closes the underlying io.ReadCloser.
func (s *StreamerV2) Close() error {
	return s.readCloser.Close()
}

type StreamThreadCreated struct {
	Thread
	streamEvent
}

type StreamThreadRunCreated struct {
	Run
	streamEvent
}

type StreamThreadRunRequiresAction struct {
	Run
	streamEvent
}

type StreamThreadRunCompleted struct {
	Run
	streamEvent
}

type StreamRunStepCompleted struct {
	RunStep
	streamEvent
}

type StreamEvent interface {
	Event() string
	JSON() json.RawMessage
}

type streamEvent struct {
	event string
	data  json.RawMessage
}

// Event returns the event name
func (s *streamEvent) Event() string {
	return s.event
}

// JSON returns the raw JSON data
func (s *streamEvent) JSON() json.RawMessage {
	return s.data
}

func (s *StreamerV2) Next() bool {
	if !s.scanner.Next() {
		return false
	}

	event := s.scanner.Scan()

	streamEvent := streamEvent{
		event: event.Event,
		data:  json.RawMessage(event.Data),
	}

	switch event.Event {
	case "thread.created":
		var thread Thread
		if err := json.Unmarshal([]byte(event.Data), &thread); err == nil {
			s.next = &StreamThreadCreated{
				Thread:      thread,
				streamEvent: streamEvent,
			}
		}
	case "thread.run.created":
		var run Run
		if err := json.Unmarshal([]byte(event.Data), &run); err == nil {
			s.next = &StreamThreadRunCreated{
				Run:         run,
				streamEvent: streamEvent,
			}
		}

	case "thread.run.requires_action":
		var run Run
		if err := json.Unmarshal([]byte(event.Data), &run); err == nil {
			s.next = &StreamThreadRunRequiresAction{
				Run:         run,
				streamEvent: streamEvent,
			}
		}
	case "thread.run.completed":
		var run Run
		if err := json.Unmarshal([]byte(event.Data), &run); err == nil {
			s.next = &StreamThreadRunCompleted{
				Run:         run,
				streamEvent: streamEvent,
			}
		}
	case "thread.message.delta":
		var delta StreamThreadMessageDelta
		if err := json.Unmarshal([]byte(event.Data), &delta); err == nil {
			delta.streamEvent = streamEvent
			s.next = &delta
		}
	case "thread.run.step.completed":
		var runStep RunStep
		if err := json.Unmarshal([]byte(event.Data), &runStep); err == nil {
			s.next = &StreamRunStepCompleted{
				RunStep:     runStep,
				streamEvent: streamEvent,
			}
		}
	case "thread.message.completed":
		var msg Message
		if err := json.Unmarshal([]byte(event.Data), &msg); err == nil {
			s.next = &StreamThreadMessageCompleted{
				Message:     msg,
				streamEvent: streamEvent,
			}
		}
	case "done":
		streamEvent.data = nil
		s.next = &StreamDone{
			streamEvent: streamEvent,
		}
	default:
		s.next = &StreamRawEvent{
			streamEvent: streamEvent,
		}
	}

	return true
}

// Read implements io.Reader of the text deltas of thread.message.delta events.
func (s *StreamerV2) Read(p []byte) (int, error) {
	// If we have data in the buffer, copy it to p first.
	if len(s.buffer) > 0 {
		n := copy(p, s.buffer)
		s.buffer = s.buffer[n:]
		return n, nil
	}

	for s.Next() {
		// Read only text deltas
		text, ok := s.MessageDeltaText()
		if !ok {
			continue
		}

		s.buffer = []byte(text)
		n := copy(p, s.buffer)
		s.buffer = s.buffer[n:]
		return n, nil
	}

	// Check for streamer error
	if err := s.Err(); err != nil {
		return 0, err
	}

	return 0, io.EOF
}

func (s *StreamerV2) Event() StreamEvent {
	return s.next
}

// Text returns text delta if the current event is a "thread.message.delta". Alias of MessageDeltaText.
func (s *StreamerV2) Text() (string, bool) {
	return s.MessageDeltaText()
}

// MessageDeltaText returns text delta if the current event is a "thread.message.delta".
func (s *StreamerV2) MessageDeltaText() (string, bool) {
	event, ok := s.next.(*StreamThreadMessageDelta)
	if !ok {
		return "", false
	}

	var text string
	for _, content := range event.Delta.Content {
		if content.Text != nil {
			// Can we return the first text we find? Does OpenAI stream ever
			// return multiple text contents in a delta?
			text += content.Text.Value
		}
	}

	return text, true
}

func (s *StreamerV2) Err() error {
	return s.scanner.Err()
}
