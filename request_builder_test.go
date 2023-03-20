package openai //nolint:testpackage // testing private field

import (
	"github.com/sashabaranov/go-openai/internal/test"

	"context"
	"errors"
	"net/http"
	"testing"
)

var (
	errTestMarshallerFailed     = errors.New("test marshaller failed")
	errTestRequestBuilderFailed = errors.New("test request builder failed")
)

type (
	failingRequestBuilder struct{}
	failingMarshaller     struct{}
)

func (*failingMarshaller) marshal(_ any) ([]byte, error) {
	return []byte{}, errTestMarshallerFailed
}

func (*failingRequestBuilder) build(_ context.Context, _, _ string, _ any) (*http.Request, error) {
	return nil, errTestRequestBuilderFailed
}

func TestRequestBuilderReturnsMarshallerErrors(t *testing.T) {
	builder := httpRequestBuilder{
		marshaller: &failingMarshaller{},
	}

	_, err := builder.build(context.Background(), "", "", struct{}{})
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}
}

func TestClientReturnsRequestBuilderErrors(t *testing.T) {
	var err error
	ts := test.NewTestServer().OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	client.requestBuilder = &failingRequestBuilder{}

	ctx := context.Background()

	_, err = client.CreateCompletion(ctx, CompletionRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateChatCompletion(ctx, ChatCompletionRequest{Model: GPT3Dot5Turbo})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateChatCompletionStream(ctx, ChatCompletionRequest{Model: GPT3Dot5Turbo})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateFineTune(ctx, FineTuneRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListFineTunes(ctx)
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CancelFineTune(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.GetFineTune(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.DeleteFineTune(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListFineTuneEvents(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.Moderations(ctx, ModerationRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.Edits(ctx, EditsRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateEmbeddings(ctx, EmbeddingRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateImage(ctx, ImageRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	err = client.DeleteFile(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.GetFile(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListFiles(ctx)
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListEngines(ctx)
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.GetEngine(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListModels(ctx)
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}
}
