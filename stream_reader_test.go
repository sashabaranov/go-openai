package openai //nolint:testpackage // testing private field

import (
	"bufio"
	"bytes"
	"errors"
	"io"
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

func TestStreamReaderOnEventCallback(t *testing.T) {
	stream := &streamReader[AssistantThreadRunStreamResponse]{
		emptyMessagesLimit: 2,
		reader: bufio.NewReader(bytes.NewReader([]byte(`
event: thread.created
data: {"id": "thread_123", "object": "thread"}
`))),
		errAccumulator: utils.NewErrorAccumulator(),
		unmarshaler:    &utils.JSONUnmarshaler{},
	}
	var response AssistantThreadRunStreamResponse
	var rawData []byte

	err := stream.On("thread.created", func(resp AssistantThreadRunStreamResponse, b []byte) {
		response = resp
		rawData = b
	})

	checks.NoError(t, err, "Stream set event callback failed")

	err = stream.On("", func(_ AssistantThreadRunStreamResponse, _ []byte) {})

	checks.HasError(t, err, "stream.On() did not return error", err.Error())

	err = stream.Run()

	checks.ErrorIs(t, err, io.EOF, "get unexpected stream error:", err.Error())

	if response.ID != "thread_123" {
		t.Fatalf("Did not retrieve the correct event id, reponse: %v", response)
	}

	if string(rawData) != `{"id": "thread_123", "object": "thread"}` {
		t.Fatalf("Did not retrieve the correct event data rawData:%s", rawData)
	}
}

func TestStreamReaderOnEventCallbackPanic(t *testing.T) {
	stream := &streamReader[AssistantThreadRunStreamResponse]{
		emptyMessagesLimit: 2,
		reader: bufio.NewReader(bytes.NewReader([]byte(`
event: thread.created
data: {"id": "thread_123", "object": "thread"}
`))),
		errAccumulator: utils.NewErrorAccumulator(),
		unmarshaler:    &utils.JSONUnmarshaler{},
	}
	var response AssistantThreadRunStreamResponse
	var rawData []byte

	err := stream.On("thread.created", func(resp AssistantThreadRunStreamResponse, b []byte) {
		response = resp
		rawData = b
		panic("event callback panic")
	})

	checks.NoError(t, err, "Stream set event callback failed")

	err = stream.On("", func(_ AssistantThreadRunStreamResponse, _ []byte) {})

	checks.HasError(t, err, "stream.On should return ErrStreamEventEmptyTopic")

	err = stream.Run()

	if errors.Is(err, io.EOF) {
		t.Fatalf("stream.Run should return ErrStreamEventCallbackPanic, but returned: %v", err)
	}

	if response.ID != "thread_123" {
		t.Fatalf("Did not retrieve the correct event id, reponse: %v", response)
	}

	if string(rawData) != `{"id": "thread_123", "object": "thread"}` {
		t.Fatalf("Did not retrieve the correct event data rawData:%s", rawData)
	}
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
		reader: bufio.NewReader(bytes.NewReader([]byte("\n"))),
		errAccumulator: &utils.DefaultErrorAccumulator{
			Buffer: &test.FailingErrorBuffer{},
		},
		unmarshaler: &utils.JSONUnmarshaler{},
	}
	_, err := stream.Recv()
	checks.ErrorIs(t, err, test.ErrTestErrorAccumulatorWriteFailed, "Did not return error when write failed", err.Error())
}
