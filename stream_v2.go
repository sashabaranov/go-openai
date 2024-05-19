package openai

import (
	"encoding/json"
	"io"
)

type StreamRawEvent struct {
	Type string
	Data json.RawMessage
}

type StreamDone struct {
}

type StreamThreadMessageDelta struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Delta  Delta  `json:"delta"`
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
		r:       rc,
		scanner: NewSSEScanner(r, false),
	}
}

type StreamerV2 struct {
	// r is only used for closing the stream
	r io.ReadCloser

	scanner *SSEScanner
	next    any

	// buffer for implementing io.Reader
	buffer []byte
}

// Close closes the underlying io.ReadCloser.
func (s *StreamerV2) Close() error {
	return s.r.Close()
}

func (s *StreamerV2) Next() bool {
	if !s.scanner.Next() {
		return false
	}

	event := s.scanner.Scan()

	switch event.Event {
	case "thread.message.delta":
		var delta StreamThreadMessageDelta
		if err := json.Unmarshal([]byte(event.Data), &delta); err == nil {
			s.next = delta
		}
	case "done":
		s.next = StreamDone{}
	default:
		s.next = StreamRawEvent{
			Type: event.Event,
			Data: json.RawMessage(event.Data),
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

func (s *StreamerV2) Event() any {
	return s.next
}

// Text returns text delta if the current event is a "thread.message.delta". Alias of MessageDeltaText.
func (s *StreamerV2) Text() (string, bool) {
	return s.MessageDeltaText()
}

// MessageDeltaText returns text delta if the current event is a "thread.message.delta".
func (s *StreamerV2) MessageDeltaText() (string, bool) {
	event, ok := s.next.(StreamThreadMessageDelta)
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
