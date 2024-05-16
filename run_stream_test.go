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

type StreamScannerV2 struct {
	scanner *SSEScanner
	next    any
}

func NewStreamScannerV2(r io.Reader) *StreamScannerV2 {
	return &StreamScannerV2{
		scanner: NewSSEScanner(r, false),
	}
}

func (s *StreamScannerV2) Next() bool {
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

func (s *StreamScannerV2) Event() any {
	return s.next
}

func (s *StreamScannerV2) Err() error {
	return s.scanner.Err()
}

func TestStreamScannerV2(t *testing.T) {
	raw := `
event: thread.message.delta
data: {"id":"msg_KFiZxHhXYQo6cGFnGjRDHSee","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"hello"}}]}}

event: done
data: [DONE]
`
	scanner := NewStreamScannerV2(strings.NewReader(raw))
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
