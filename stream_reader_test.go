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

// TestStreamReaderReturnsReadErrWhenErrorResponseHasNilError tests that when
// unmarshalError returns an ErrorResponse with a nil Error field (e.g., when
// unmarshaling an empty JSON object "{}"), the original read error is returned
// instead of "error, <nil>". This fixes issue #1060.
func TestStreamReaderReturnsReadErrWhenErrorResponseHasNilError(t *testing.T) {
	// Create a stream that will return io.EOF on read (simulating context cancellation)
	// with an error accumulator that contains an empty JSON object
	stream := &streamReader[ChatCompletionStreamResponse]{
		reader:         bufio.NewReader(bytes.NewReader([]byte{})), // empty reader returns io.EOF
		errAccumulator: utils.NewErrorAccumulator(),
		unmarshaler:    &utils.JSONUnmarshaler{},
	}

	// Write an empty JSON object to the error accumulator
	// This will unmarshal to ErrorResponse{Error: nil}
	err := stream.errAccumulator.Write([]byte("{}"))
	if err != nil {
		t.Fatalf("Failed to write to error accumulator: %v", err)
	}

	// Call processLines which should return io.EOF, not "error, <nil>"
	_, err = stream.processLines()
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
	if !errors.Is(err, io.EOF) {
		t.Fatalf("Expected io.EOF but got: %v", err)
	}
}
