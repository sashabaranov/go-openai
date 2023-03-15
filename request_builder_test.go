package openai //nolint:testpackage // testing private field

import (
	"github.com/sashabaranov/go-openai/internal/test"

	"context"
	"errors"
	"net/http"
	"testing"
)

type failingRequestBuilder struct{}

var errTestRequestBuilderFailed = errors.New("test request builder failed")

func (*failingRequestBuilder) build(ctx context.Context, method, url string, requset any) (*http.Request, error) {
	return nil, errTestRequestBuilderFailed
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

	_, err = client.CreateChatCompletionStream(ctx, ChatCompletionRequest{})
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
}
