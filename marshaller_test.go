package openai //nolint:testpackage // testing private field

import (
	"github.com/sashabaranov/go-openai/internal/test"

	"context"
	"errors"
	"testing"
)

type failingMarshaller struct{}

var errTestMarshallerFailed = errors.New("test marshaller failed")

func (jm *failingMarshaller) marshal(value any) ([]byte, error) {
	return []byte{}, errTestMarshallerFailed
}

func TestClientReturnMarshallerErrors(t *testing.T) {
	var err error
	ts := test.NewTestServer().OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	client.marshaller = &failingMarshaller{}

	ctx := context.Background()

	_, err = client.CreateCompletion(ctx, CompletionRequest{})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}

	_, err = client.CreateChatCompletion(ctx, ChatCompletionRequest{Model: GPT3Dot5Turbo})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}

	_, err = client.CreateChatCompletionStream(ctx, ChatCompletionRequest{})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}

	_, err = client.CreateFineTune(ctx, FineTuneRequest{})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}

	_, err = client.Moderations(ctx, ModerationRequest{})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}

	_, err = client.Edits(ctx, EditsRequest{})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}

	_, err = client.CreateEmbeddings(ctx, EmbeddingRequest{})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}

	_, err = client.CreateImage(ctx, ImageRequest{})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}
}
