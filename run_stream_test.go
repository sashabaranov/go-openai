package openai

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
)

type StreamRawEvent struct {
	Type string
	Data json.RawMessage
}

type StreamDone struct {
	Data string // [DONE]
}

// Define StreamThreadMessageDelta
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

// StreamTextReader wraps StreamerV2 to implement io.Reader.
type StreamTextReader struct {
	streamer *StreamerV2
	buffer   []byte
}

// NewStreamTextReader initializes and returns a new StreamTextReader.
func NewStreamTextReader(streamer *StreamerV2) *StreamTextReader {
	return &StreamTextReader{
		streamer: streamer,
	}
}

// Read implements the io.Reader interface.
func (r *StreamTextReader) Read(p []byte) (int, error) {
	// If we have data in the buffer, copy it to p first.
	if len(r.buffer) > 0 {
		n := copy(p, r.buffer)
		r.buffer = r.buffer[n:]
		return n, nil
	}

	for r.streamer.Next() {
		event := r.streamer.Event()
		switch e := event.(type) {
		case StreamThreadMessageDelta:
			// Check if the event contains text content.
			for _, content := range e.Delta.Content {
				if content.Text != nil {
					r.buffer = []byte(content.Text.Value)
					n := copy(p, r.buffer)
					r.buffer = r.buffer[n:]
					return n, nil
				}
			}
		case StreamDone:
			return 0, io.EOF
		}
	}

	// If we reach here, there are no more events.
	return 0, io.EOF
}

func NewStreamerV2(r io.Reader) *StreamerV2 {
	return &StreamerV2{
		scanner: NewSSEScanner(r, false),
	}
}

type StreamerV2 struct {
	scanner *SSEScanner
	next    any
}

func (s *StreamerV2) Next() bool {
	if s.scanner.Next() {
		event := s.scanner.Scan()
		if event != nil {
			switch event.Event {
			case "thread.message.delta":
				var delta StreamThreadMessageDelta
				if err := json.Unmarshal([]byte(event.Data), &delta); err == nil {
					s.next = delta
					return true
				}
			case "done":
				s.next = StreamDone{Data: "DONE"}
				return true
			default:
				s.next = StreamRawEvent{Data: json.RawMessage(event.Data)}
			}
		}
	}
	return false
}

func (s *StreamerV2) Event() any {
	return s.next
}

func (s *StreamerV2) Err() error {
	return s.scanner.Err()
}

func TestNewStreamTextReader(t *testing.T) {
	raw := `
event: thread.message.delta
data: {"id":"msg_KFiZxHhXYQo6cGFnGjRDHSee","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"hello"}}]}}

event: thread.message.delta
data: {"id":"msg_KFiZxHhXYQo6cGFnGjRDHSee","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"world"}}]}}

event: done
data: [DONE]
`
	scanner := NewStreamerV2(strings.NewReader(raw))
	reader := NewStreamTextReader(scanner)

	expected := "helloworld"
	buffer := make([]byte, len(expected))
	n, err := reader.Read(buffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("hello") {
		t.Fatalf("expected to read %d bytes, read %d bytes", len("hello"), n)
	}
	if string(buffer[:n]) != "hello" {
		t.Fatalf("expected %q, got %q", "hello", string(buffer[:n]))
	}

	n, err = reader.Read(buffer[n:])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("world") {
		t.Fatalf("expected to read %d bytes, read %d bytes", len("world"), n)
	}
	if string(buffer[:len(expected)]) != expected {
		t.Fatalf("expected %q, got %q", expected, string(buffer[:len(expected)]))
	}

	n, err = reader.Read(buffer)
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}
	if n != 0 {
		t.Fatalf("expected to read 0 bytes, read %d bytes", n)
	}
}

func TestStreamScannerV2(t *testing.T) {
	raw := `
event: thread.message.delta
data: {"id":"msg_KFiZxHhXYQo6cGFnGjRDHSee","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"hello"}}]}}

event: done
data: [DONE]
`
	scanner := NewStreamerV2(strings.NewReader(raw))
	var events []any

	for scanner.Next() {
		event := scanner.Event()
		events = append(events, event)
	}

	expectedValues := []any{
		StreamThreadMessageDelta{
			ID:     "msg_KFiZxHhXYQo6cGFnGjRDHSee",
			Object: "thread.message.delta",
			Delta: Delta{
				Content: []DeltaContent{
					{
						Index: 0,
						Type:  "text",
						Text: &DeltaText{
							Value: "hello",
						},
					},
				},
			},
		},
		StreamDone{Data: "DONE"},
	}

	if len(events) != len(expectedValues) {
		t.Fatalf("Expected %d events but got %d", len(expectedValues), len(events))
	}

	for i, event := range events {
		expectedValue := expectedValues[i]
		if !reflect.DeepEqual(event, expectedValue) {
			t.Errorf("Expected %v but got %v", expectedValue, event)
		}
	}
}

func TestStreamThreadMessageDeltaJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectType  string
		expectValue interface{}
	}{
		{
			name:        "DeltaContent with Text",
			jsonData:    `{"index":0,"type":"text","text":{"value":"hello"}}`,
			expectType:  "text",
			expectValue: &DeltaText{Value: "hello"},
		},
		{
			name:        "DeltaContent with ImageFile",
			jsonData:    `{"index":1,"type":"image_file","image_file":{"file_id":"file123","detail":"An image"}}`,
			expectType:  "image_file",
			expectValue: &DeltaImageFile{FileID: "file123", Detail: "An image"},
		},
		{
			name:        "DeltaContent with ImageURL",
			jsonData:    `{"index":2,"type":"image_url","image_url":{"url":"https://example.com/image.jpg","detail":"low"}}`,
			expectType:  "image_url",
			expectValue: &DeltaImageURL{URL: "https://example.com/image.jpg", Detail: "low"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var content DeltaContent
			err := json.Unmarshal([]byte(tt.jsonData), &content)
			if err != nil {
				t.Fatalf("Error unmarshalling JSON: %v", err)
			}

			if content.Type != tt.expectType {
				t.Errorf("Expected Type to be '%s', got %s", tt.expectType, content.Type)
			}

			var actualValue interface{}
			switch tt.expectType {
			case "text":
				actualValue = content.Text
			case "image_file":
				actualValue = content.ImageFile
			case "image_url":
				actualValue = content.ImageURL
			default:
				t.Fatalf("Unexpected type: %s", tt.expectType)
			}

			if !reflect.DeepEqual(actualValue, tt.expectValue) {
				t.Errorf("Expected value to be '%v', got %v", tt.expectValue, actualValue)
			}
		})
	}
}
