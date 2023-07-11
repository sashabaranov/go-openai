//go:build integration

package openai_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
	"github.com/sashabaranov/go-openai/jsonschema"
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

	_, err = c.CreateChatCompletion(
		context.Background(),
		ChatCompletionRequest{
			Model: GPT3Dot5Turbo,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Content: "What is the weather like in Boston?",
				},
			},
			Functions: []FunctionDefinition{{
				Name: "get_current_weather",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"location": {
							Type:        jsonschema.String,
							Description: "The city and state, e.g. San Francisco, CA",
						},
						"unit": {
							Type: jsonschema.String,
							Enum: []string{"celsius", "fahrenheit"},
						},
					},
					Required: []string{"location"},
				},
			}},
		},
	)
	checks.NoError(t, err, "CreateChatCompletion (with functions) returned error")
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
	checks.HasError(t, err, "ListEngines should fail with an invalid key")

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Error is not an APIError: %+v", err)
	}

	if apiErr.HTTPStatusCode != 401 {
		t.Fatalf("Unexpected API error status code: %d", apiErr.HTTPStatusCode)
	}

	switch v := apiErr.Code.(type) {
	case string:
		if v != "invalid_api_key" {
			t.Fatalf("Unexpected API error code: %s", v)
		}
	default:
		t.Fatalf("Unexpected API error code type: %T", v)
	}

	if apiErr.Error() == "" {
		t.Fatal("Empty error message occurred")
	}
}
