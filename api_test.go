package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"context"
	"errors"
	"io"
	"os"
	"testing"
)

func TestAPI(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	var err error
	c := NewClient(apiToken)
	ctx := context.Background()
	_, err = c.ListEngines(ctx)
	checks.NoError(t, err, "ListEngines error")

	_, err = c.GetEngine(ctx, "davinci")
	checks.NoError(t, err, "GetEngine error")

	fileRes, err := c.ListFiles(ctx)
	checks.NoError(t, err, "ListFiles error")

	if len(fileRes.Files) > 0 {
		_, err = c.GetFile(ctx, fileRes.Files[0].ID)
		checks.NoError(t, err, "GetFile error")
	} // else skip

	embeddingReq := EmbeddingRequest{
		Input: []string{
			"The food was delicious and the waiter",
			"Other examples of embedding request",
		},
		Model: AdaSearchQuery,
	}
	_, err = c.CreateEmbeddings(ctx, embeddingReq)
	checks.NoError(t, err, "Embedding error")

	_, err = c.CreateChatCompletion(
		ctx,
		ChatCompletionRequest{
			Model: GPT3Dot5Turbo,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	checks.NoError(t, err, "CreateChatCompletion (without name) returned error")

	_, err = c.CreateChatCompletion(
		ctx,
		ChatCompletionRequest{
			Model: GPT3Dot5Turbo,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Name:    "John_Doe",
					Content: "Hello!",
				},
			},
		},
	)
	checks.NoError(t, err, "CreateChatCompletion (with name) returned error")

	stream, err := c.CreateCompletionStream(ctx, CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     GPT3Ada,
		MaxTokens: 5,
		Stream:    true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	counter := 0
	for {
		_, err = stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Errorf("Stream error: %v", err)
		} else {
			counter++
		}
	}
	if counter == 0 {
		t.Error("Stream did not return any responses")
	}
}

func TestAPIError(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	var err error
	c := NewClient(apiToken + "_invalid")
	ctx := context.Background()
	_, err = c.ListEngines(ctx)
	checks.NoError(t, err, "ListEngines did not fail")

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Error is not an APIError: %+v", err)
	}

	if apiErr.StatusCode != 401 {
		t.Fatalf("Unexpected API error status code: %d", apiErr.StatusCode)
	}
	if *apiErr.Code != "invalid_api_key" {
		t.Fatalf("Unexpected API error code: %s", *apiErr.Code)
	}
	if apiErr.Error() == "" {
		t.Fatal("Empty error message occurred")
	}
}

func TestRequestError(t *testing.T) {
	var err error

	config := DefaultConfig("dummy")
	config.BaseURL = "https://httpbin.org/status/418?"
	c := NewClientWithConfig(config)
	ctx := context.Background()
	_, err = c.ListEngines(ctx)
	checks.HasError(t, err, "ListEngines did not fail")

	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("Error is not a RequestError: %+v", err)
	}

	if reqErr.StatusCode != 418 {
		t.Fatalf("Unexpected request error status code: %d", reqErr.StatusCode)
	}

	if reqErr.Unwrap() == nil {
		t.Fatalf("Empty request error occurred")
	}
}

// numTokens Returns the number of GPT-3 encoded tokens in the given text.
// This function approximates based on the rule of thumb stated by OpenAI:
// https://beta.openai.com/tokenizer
//
// TODO: implement an actual tokenizer for GPT-3 and Codex (once available)
func numTokens(s string) int {
	return int(float32(len(s)) / 4)
}
