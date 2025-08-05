package openai //nolint:testpackage // testing private field

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	utils "github.com/sashabaranov/go-openai/internal"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

var errTestUnmarshalerFailed = errors.New("test unmarshaler failed")

type failingUnMarshaller struct{}

func (*failingUnMarshaller) Unmarshal(_ []byte, _ any) error {
	return errTestUnmarshalerFailed
}

func TestStreamReaderReturnsUnmarshalerErrors(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		errAccumulator: utils.NewErrorAccumulator(),
		unmarshaler:    &failingUnMarshaller{},
	}

	respErr := stream.unmarshalError()
	if respErr != nil {
		t.Fatalf("Did not return nil with empty buffer: %v", respErr)
	}

	err := stream.errAccumulator.Write([]byte("{"))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	respErr = stream.unmarshalError()
	if respErr != nil {
		t.Fatalf("Did not return nil when unmarshaler failed: %v", respErr)
	}
}

func TestStreamReaderReturnsErrTooManyEmptyStreamMessages(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		emptyMessagesLimit: 3,
		reader:             bufio.NewReader(bytes.NewReader([]byte("\n\n\n\n"))),
		errAccumulator:     utils.NewErrorAccumulator(),
		unmarshaler:        &utils.JSONUnmarshaler{},
	}
	_, err := stream.Recv()
	checks.ErrorIs(t, err, ErrTooManyEmptyStreamMessages, "Did not return error when recv failed", err.Error())
}

func TestStreamReaderReturnsErrTestErrorAccumulatorWriteFailed(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		reader: bufio.NewReader(bytes.NewReader([]byte("data: {\"error\": {\"message\": \"test error\"}}\n"))),
		errAccumulator: &utils.DefaultErrorAccumulator{
			Buffer: &test.FailingErrorBuffer{},
		},
		unmarshaler:        &utils.JSONUnmarshaler{},
		emptyMessagesLimit: 5,
	}
	_, err := stream.Recv()
	checks.ErrorIs(t, err, test.ErrTestErrorAccumulatorWriteFailed, "Did not return error when write failed", err.Error())
}

func TestStreamReaderRecvRaw(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		reader: bufio.NewReader(bytes.NewReader([]byte("data: {\"key\": \"value\"}\n"))),
	}
	rawLine, err := stream.RecvRaw()
	if err != nil {
		t.Fatalf("Did not return raw line: %v", err)
	}
	if !bytes.Equal(rawLine, []byte("{\"key\": \"value\"}")) {
		t.Fatalf("Did not return raw line: %v", string(rawLine))
	}
}

func TestStreamReaderParsesErrorEvents(t *testing.T) {
	// Test case simulating Groq's error event format
	errorEvent := `event: error
data: {"error":{"message":"Invalid tool_call: tool \"name_unknown\" does not exist.","type":"invalid_request_error","code":"invalid_tool_call"}}

`
	stream := &streamReader[ChatCompletionStreamResponse]{
		reader:             bufio.NewReader(bytes.NewReader([]byte(errorEvent))),
		errAccumulator:     utils.NewErrorAccumulator(),
		unmarshaler:        &utils.JSONUnmarshaler{},
		emptyMessagesLimit: 5,
	}

	// Process the error event
	_, err := stream.Recv()
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	// Verify it's an APIError
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError type but got %T: %v", err, err)
	}

	// Verify the error fields are correctly parsed
	if apiErr.Message != "Invalid tool_call: tool \"name_unknown\" does not exist." {
		t.Errorf("Unexpected error message: %s", apiErr.Message)
	}
	if apiErr.Type != "invalid_request_error" {
		t.Errorf("Unexpected error type: %s", apiErr.Type)
	}
	if apiErr.Code != "invalid_tool_call" {
		t.Errorf("Unexpected error code: %v", apiErr.Code)
	}
}

func TestStreamReaderHandlesErrorEventWithExtraData(t *testing.T) {
	// Test case with error event followed by more data
	errorEvent := `data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"Hello"}}]}
event: error
data: {"error":{"message":"Stream interrupted","type":"server_error"}}
data: [DONE]
`
	stream := &streamReader[ChatCompletionStreamResponse]{
		reader:             bufio.NewReader(bytes.NewReader([]byte(errorEvent))),
		errAccumulator:     utils.NewErrorAccumulator(),
		unmarshaler:        &utils.JSONUnmarshaler{},
		emptyMessagesLimit: 5,
	}

	// First recv should return the chat completion
	resp, err := stream.Recv()
	if err != nil {
		t.Fatalf("First recv failed: %v", err)
	}
	if resp.ID != "chatcmpl-123" {
		t.Errorf("Unexpected response ID: %s", resp.ID)
	}

	// Second recv should return the error
	_, err = stream.Recv()
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	// Verify it's an APIError
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError type but got %T: %v", err, err)
	}

	if apiErr.Message != "Stream interrupted" {
		t.Errorf("Unexpected error message: %s", apiErr.Message)
	}
}

func TestStreamReaderResetsErrorAccumulator(t *testing.T) {
	// Test that error accumulator is reset after processing an error
	multipleErrors := `event: error
data: {"error":{"message":"First error","type":"error_type_1"}}

event: error  
data: {"error":{"message":"Second error","type":"error_type_2"}}
`
	stream := &streamReader[ChatCompletionStreamResponse]{
		reader:             bufio.NewReader(bytes.NewReader([]byte(multipleErrors))),
		errAccumulator:     utils.NewErrorAccumulator(),
		unmarshaler:        &utils.JSONUnmarshaler{},
		emptyMessagesLimit: 5,
	}

	// First recv should return the first error
	_, err1 := stream.Recv()
	if err1 == nil {
		t.Fatal("Expected first error but got nil")
	}
	apiErr1, ok := err1.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError type but got %T: %v", err1, err1)
	}
	if apiErr1.Message != "First error" {
		t.Errorf("Unexpected first error message: %s", apiErr1.Message)
	}

	// Second recv should return the second error (not a concatenation)
	_, err2 := stream.Recv()
	if err2 == nil {
		t.Fatal("Expected second error but got nil")
	}
	apiErr2, ok := err2.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError type but got %T: %v", err2, err2)
	}
	if apiErr2.Message != "Second error" {
		t.Errorf("Unexpected second error message: %s", apiErr2.Message)
	}
	if apiErr2.Type != "error_type_2" {
		t.Errorf("Unexpected second error type: %s", apiErr2.Type)
	}
}
