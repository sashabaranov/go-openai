package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

var (
	errTestUnmarshalerFailed           = errors.New("test unmarshaler failed")
	errTestErrorAccumulatorWriteFailed = errors.New("test error accumulator failed")
)

type (
	failingUnMarshaller struct{}
	failingErrorBuffer  struct{}
)

func (b *failingErrorBuffer) Write(_ []byte) (n int, err error) {
	return 0, errTestErrorAccumulatorWriteFailed
}

func (b *failingErrorBuffer) Len() int {
	return 0
}

func (b *failingErrorBuffer) Bytes() []byte {
	return []byte{}
}

func (*failingUnMarshaller) unmarshal(_ []byte, _ any) error {
	return errTestUnmarshalerFailed
}

func TestErrorAccumulatorReturnsUnmarshalerErrors(t *testing.T) {
	accumulator := &defaultErrorAccumulator{
		buffer:      &bytes.Buffer{},
		unmarshaler: &failingUnMarshaller{},
	}

	respErr := accumulator.unmarshalError()
	if respErr != nil {
		t.Fatalf("Did not return nil with empty buffer: %v", respErr)
	}

	err := accumulator.write([]byte("{"))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	respErr = accumulator.unmarshalError()
	if respErr != nil {
		t.Fatalf("Did not return nil when unmarshaler failed: %v", respErr)
	}
}

func TestErrorByteWriteErrors(t *testing.T) {
	accumulator := &defaultErrorAccumulator{
		buffer:      &failingErrorBuffer{},
		unmarshaler: &jsonUnmarshaler{},
	}
	err := accumulator.write([]byte("{"))
	if !errors.Is(err, errTestErrorAccumulatorWriteFailed) {
		t.Fatalf("Did not return error when write failed: %v", err)
	}
}

func TestErrorAccumulatorWriteErrors(t *testing.T) {
	var err error
	ts := test.NewTestServer().OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)

	ctx := context.Background()

	stream, err := client.CreateChatCompletionStream(ctx, ChatCompletionRequest{})
	checks.NoError(t, err)

	stream.errAccumulator = &defaultErrorAccumulator{
		buffer:      &failingErrorBuffer{},
		unmarshaler: &jsonUnmarshaler{},
	}

	_, err = stream.Recv()
	checks.ErrorIs(t, err, errTestErrorAccumulatorWriteFailed, "Did not return error when write failed", err.Error())
}
