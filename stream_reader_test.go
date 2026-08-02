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

// Regression for #1060. When the read aborts (context cancel, network
// reset, etc) and whatever was buffered in errAccumulator happens to
// parse as an ErrorResponse with a nil Error field, the stream used to
// surface a useless "error, <nil>" instead of the real read error.
func TestStreamReaderPropagatesReadErrorWhenErrorResponseIsEmpty(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		reader:         bufio.NewReader(bytes.NewReader([]byte(""))),
		errAccumulator: utils.NewErrorAccumulator(),
		unmarshaler:    &utils.JSONUnmarshaler{},
	}
	// Prime the accumulator with a payload that unmarshals successfully
	// into ErrorResponse{} but leaves the Error pointer nil. This mirrors
	// what we end up with after a partial SSE frame plus EOF.
	if err := stream.errAccumulator.Write([]byte("{}")); err != nil {
		t.Fatalf("write to accumulator: %v", err)
	}

	_, err := stream.Recv()
	checks.ErrorIs(t, err, io.EOF, "expected the underlying io.EOF to propagate")
}
