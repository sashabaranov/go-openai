package openai_test

import (
	. "github.com/sashabaranov/go-openai"

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
	if err != nil {
		t.Fatalf("ListEngines error: %v", err)
	}

	_, err = c.GetEngine(ctx, "davinci")
	if err != nil {
		t.Fatalf("GetEngine error: %v", err)
	}

	fileRes, err := c.ListFiles(ctx)
	if err != nil {
		t.Fatalf("ListFiles error: %v", err)
	}

	if len(fileRes.Files) > 0 {
		_, err = c.GetFile(ctx, fileRes.Files[0].ID)
		if err != nil {
			t.Fatalf("GetFile error: %v", err)
		}
	} // else skip

	embeddingReq := EmbeddingRequest{
		Input: []string{
			"The food was delicious and the waiter",
			"Other examples of embedding request",
		},
		Model: AdaSearchQuery,
	}
	_, err = c.CreateEmbeddings(ctx, embeddingReq)
	if err != nil {
		t.Fatalf("Embedding error: %v", err)
	}

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

	if err != nil {
		t.Errorf("CreateChatCompletion (without name) returned error: %v", err)
	}

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

	if err != nil {
		t.Errorf("CreateChatCompletion (with name) returned error: %v", err)
	}

	stream, err := c.CreateCompletionStream(ctx, CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     GPT3Ada,
		MaxTokens: 5,
		Stream:    true,
	})
	if err != nil {
		t.Errorf("CreateCompletionStream returned error: %v", err)
	}
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
	if err == nil {
		t.Fatal("ListEngines did not fail")
	}

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
}

func TestRequestError(t *testing.T) {
	var err error

	config := DefaultConfig("dummy")
	config.BaseURL = "https://httpbin.org/status/418?"
	c := NewClientWithConfig(config)
	ctx := context.Background()
	_, err = c.ListEngines(ctx)
	if err == nil {
		t.Fatal("ListEngines request did not fail")
	}

	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("Error is not a RequestError: %+v", err)
	}

	if reqErr.StatusCode != 418 {
		t.Fatalf("Unexpected request error status code: %d", reqErr.StatusCode)
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
