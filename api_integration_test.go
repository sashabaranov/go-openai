//go:build integration

package openai_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestAPI(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	var err error
	c := openai.NewClient(apiToken)
	ctx := context.Background()
	_, err = c.ListEngines(ctx)
	checks.NoError(t, err, "ListEngines error")

	_, err = c.GetEngine(ctx, openai.GPT3Davinci002)
	checks.NoError(t, err, "GetEngine error")

	fileRes, err := c.ListFiles(ctx)
	checks.NoError(t, err, "ListFiles error")

	if len(fileRes.Files) > 0 {
		_, err = c.GetFile(ctx, fileRes.Files[0].ID)
		checks.NoError(t, err, "GetFile error")
	} // else skip

	embeddingReq := openai.EmbeddingRequest{
		Input: []string{
			"The food was delicious and the waiter",
			"Other examples of embedding request",
		},
		Model: openai.AdaEmbeddingV2,
	}
	_, err = c.CreateEmbeddings(ctx, embeddingReq)
	checks.NoError(t, err, "Embedding error")

	_, err = c.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	checks.NoError(t, err, "CreateChatCompletion (without name) returned error")

	_, err = c.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Name:    "John_Doe",
					Content: "Hello!",
				},
			},
		},
	)
	checks.NoError(t, err, "CreateChatCompletion (with name) returned error")

	_, err = c.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "What is the weather like in Boston?",
				},
			},
			Functions: []openai.FunctionDefinition{{
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

func TestCompletionStream(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	c := openai.NewClient(apiToken)
	ctx := context.Background()

	stream, err := c.CreateCompletionStream(ctx, openai.CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     openai.GPT3Babbage002,
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

func TestBatchAPI(t *testing.T) {
	ctx := context.Background()
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}
	var err error
	c := openai.NewClient(apiToken)

	req := openai.CreateBatchWithUploadFileRequest{
		Endpoint:         openai.BatchEndpointChatCompletions,
		CompletionWindow: "24h",
	}
	for i := 0; i < 5; i++ {
		req.AddChatCompletion(fmt.Sprintf("req-%d", i), openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("What is the square of %d?", i+1),
				},
			},
		})
	}
	_, err = c.CreateBatchWithUploadFile(ctx, req)
	checks.NoError(t, err, "CreateBatchWithUploadFile error")

	var chatCompletions = make([]openai.BatchChatCompletion, 5)
	for i := 0; i < 5; i++ {
		chatCompletions[i] = openai.BatchChatCompletion{
			CustomID: fmt.Sprintf("req-%d", i),
			ChatCompletion: openai.ChatCompletionRequest{
				Model: openai.GPT4oMini,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("What is the square of %d?", i+1),
					},
				},
			},
		}
	}
	_, err = c.CreateBatchWithChatCompletions(ctx, openai.CreateBatchWithChatCompletionsRequest{
		ChatCompletions: chatCompletions,
	})
	checks.NoError(t, err, "CreateBatchWithChatCompletions error")

	var embeddings = make([]openai.BatchEmbedding, 3)
	for i := 0; i < 3; i++ {
		embeddings[i] = openai.BatchEmbedding{
			CustomID: fmt.Sprintf("req-%d", i),
			Embedding: openai.EmbeddingRequest{
				Input:          "The food was delicious and the waiter...",
				Model:          openai.AdaEmbeddingV2,
				EncodingFormat: openai.EmbeddingEncodingFormatFloat,
			},
		}
	}
	_, err = c.CreateBatchWithEmbeddings(ctx, openai.CreateBatchWithEmbeddingsRequest{
		Embeddings: embeddings,
	})
	checks.NoError(t, err, "CreateBatchWithEmbeddings error")

	_, err = c.ListBatch(ctx, nil, nil)
	checks.NoError(t, err, "ListBatch error")
}

func TestAPIError(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	var err error
	c := openai.NewClient(apiToken + "_invalid")
	ctx := context.Background()
	_, err = c.ListEngines(ctx)
	checks.HasError(t, err, "ListEngines should fail with an invalid key")

	var apiErr *openai.APIError
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
