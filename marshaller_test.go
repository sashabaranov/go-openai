package openai //nolint:testpackage // testing private field

import (
	"context"
	"errors"
	"testing"
)

type failingMarshaller struct{}

var errTestMarshallerFailed = errors.New("test marshaller failed")

func (jm *failingMarshaller) marshal(value any) ([]byte, error) {
	return []byte{}, errTestMarshallerFailed
}

func TestRequestBuilderReturnsMarshallerErrors(t *testing.T) {
	builder := httpRequestBuilder{
		marshaller: &failingMarshaller{},
	}

	_, err := builder.build(context.Background(), "", "", nil)
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}
}
