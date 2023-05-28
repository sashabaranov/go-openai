package openai //nolint:testpackage // testing private field

import (
	"context"
	"errors"
	"testing"
)

var errTestMarshallerFailed = errors.New("test marshaller failed")

type failingMarshaller struct{}

func (*failingMarshaller) Marshal(_ any) ([]byte, error) {
	return []byte{}, errTestMarshallerFailed
}

func TestRequestBuilderReturnsMarshallerErrors(t *testing.T) {
	builder := HTTPRequestBuilder{
		marshaller: &failingMarshaller{},
	}

	_, err := builder.Build(context.Background(), "", "", struct{}{})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}
}
