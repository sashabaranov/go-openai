//nolint:lll
package openai_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
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
	reader := openai.NewStreamerV2(strings.NewReader(raw))

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
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
	if n != 0 {
		t.Fatalf("expected to read 0 bytes, read %d bytes", n)
	}
}

type TestCase struct {
	Event string
	Data  string
}

func constructStreamInput(testCases []TestCase) io.Reader {
	var sb bytes.Buffer
	for _, tc := range testCases {
		sb.WriteString("event: ")
		sb.WriteString(tc.Event)
		sb.WriteString("\n")
		sb.WriteString("data: ")
		sb.WriteString(tc.Data)
		sb.WriteString("\n\n")
	}
	return &sb
}

func jsonEqual[T any](t *testing.T, data []byte, expected T) error {
	var obj T
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	if !reflect.DeepEqual(obj, expected) {
		t.Fatalf("Expected %v, but got %v", expected, obj)
	}

	return nil
}

func TestStreamerV2(t *testing.T) {
	testCases := []TestCase{
		{
			Event: "thread.created",
			Data:  `{"id":"thread_vMWb8sJ14upXpPO2VbRpGTYD","object":"thread","created_at":1715864046,"metadata":{},"tool_resources":{"code_interpreter":{"file_ids":[]}}}`,
		},
		{
			Event: "thread.run.created",
			Data:  `{"id":"run_ojU7pVxtTIaa4l1GgRmHVSbK","object":"thread.run","created_at":1715864046,"assistant_id":"asst_7xUrZ16RBU2BpaUOzLnc9HsD","thread_id":"thread_vMWb8sJ14upXpPO2VbRpGTYD","status":"queued","started_at":null,"expires_at":1715864646,"cancelled_at":null,"failed_at":null,"completed_at":null,"required_action":null,"last_error":null,"model":"gpt-3.5-turbo","instructions":null,"tools":[],"tool_resources":{"code_interpreter":{"file_ids":[]}},"metadata":{},"temperature":1.0,"top_p":1.0,"max_completion_tokens":null,"max_prompt_tokens":null,"truncation_strategy":{"type":"auto","last_messages":null},"incomplete_details":null,"usage":null,"response_format":"auto","tool_choice":"auto"}`,
		},
		{
			Event: "thread.message.delta",
			Data:  `{"id":"msg_KFiZxHhXYQo6cGFnGjRDHSee","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"hello"}}]}}`,
		},
		{
			Event: "done",
			Data:  "[DONE]",
		},
	}

	streamer := openai.NewStreamerV2(constructStreamInput(testCases))

	for _, tc := range testCases {
		if !streamer.Next() {
			t.Fatal("Expected Next() to return true, but got false")
		}

		event := streamer.Event()

		if event.Event() != tc.Event {
			t.Fatalf("Expected event type to be %s, but got %s", tc.Event, event.Event())
		}

		if tc.Event != "done" {
			// compare the json data
			jsondata := event.JSON()
			if string(jsondata) != tc.Data {
				t.Fatalf("Expected JSON data to be %s, but got %s", tc.Data, string(jsondata))
			}
		}

		switch event := event.(type) {
		case *openai.StreamThreadCreated:
			jsonEqual(t, []byte(tc.Data), event.Thread)
		case *openai.StreamThreadRunCreated:
			jsonEqual(t, []byte(tc.Data), event.Run)
		case *openai.StreamThreadMessageDelta:
			fmt.Println(event)

			// reinitialize the delta object to avoid comparing the hidden streamEvent fields
			delta := openai.StreamThreadMessageDelta{
				ID:     event.ID,
				Object: event.Object,
				Delta:  event.Delta,
			}

			jsonEqual(t, []byte(tc.Data), delta)
		case *openai.StreamDone:
			if event.JSON() != nil {
				t.Fatalf("Expected JSON data to be nil, but got %s", string(event.JSON()))
			}
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
			expectValue: &openai.DeltaText{Value: "hello"},
		},
		{
			name:        "DeltaContent with ImageFile",
			jsonData:    `{"index":1,"type":"image_file","image_file":{"file_id":"file123","detail":"An image"}}`,
			expectType:  "image_file",
			expectValue: &openai.DeltaImageFile{FileID: "file123", Detail: "An image"},
		},
		{
			name:        "DeltaContent with ImageURL",
			jsonData:    `{"index":2,"type":"image_url","image_url":{"url":"https://example.com/image.jpg","detail":"low"}}`,
			expectType:  "image_url",
			expectValue: &openai.DeltaImageURL{URL: "https://example.com/image.jpg", Detail: "low"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var content openai.DeltaContent
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
