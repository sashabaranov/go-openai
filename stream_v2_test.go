package openai

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestNewStreamTextReader(t *testing.T) {
	raw := `
event: thread.message.delta
data: {"id":"msg_KFiZxHhXYQo6cGFnGjRDHSee","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"hello"}}]}}

event: thread.message.delta
data: {"id":"msg_KFiZxHhXYQo6cGFnGjRDHSee","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"world"}}]}}

event: done
data: [DONE]
`
	reader := NewStreamerV2(strings.NewReader(raw))

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
	raw := `event: thread.created
data: {"id":"thread_vMWb8sJ14upXpPO2VbRpGTYD","object":"thread","created_at":1715864046,"metadata":{},"tool_resources":{"code_interpreter":{"file_ids":[]}}}

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
		StreamRawEvent{
			Type: "thread.created",
			Data: json.RawMessage(`{"id":"thread_vMWb8sJ14upXpPO2VbRpGTYD","object":"thread","created_at":1715864046,"metadata":{},"tool_resources":{"code_interpreter":{"file_ids":[]}}}`),
		},
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
		StreamDone{},
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
