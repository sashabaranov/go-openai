package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"errors"
	"testing"
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

func (*failingUnMarshaller) Unmarshal(_ []byte, _ any) error {
	return errTestUnmarshalerFailed
}

func TestErrorAccumulatorBytes(t *testing.T) {
	accumulator := &defaultErrorAccumulator{
		buffer: &bytes.Buffer{},
	}

	errBytes := accumulator.bytes()
	if len(errBytes) != 0 {
		t.Fatalf("Did not return nil with empty bytes: %s", string(errBytes))
	}

	err := accumulator.write([]byte("{}"))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	errBytes = accumulator.bytes()
	if len(errBytes) == 0 {
		t.Fatalf("Did not return error bytes when has error: %s", string(errBytes))
	}
}

func TestErrorByteWriteErrors(t *testing.T) {
	accumulator := &defaultErrorAccumulator{
		buffer: &failingErrorBuffer{},
	}
	err := accumulator.write([]byte("{"))
	if !errors.Is(err, errTestErrorAccumulatorWriteFailed) {
		t.Fatalf("Did not return error when write failed: %v", err)
	}
}
