package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
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

func TestAPIErrorUnmarshalJSONInteger(t *testing.T) {
	var apiErr APIError
	response := `{"code":418,"message":"I'm a teapot","param":"prompt","type":"teapot_error"}`
	err := json.Unmarshal([]byte(response), &apiErr)
	checks.NoError(t, err, "Unexpected Unmarshal API response error")

	switch v := apiErr.Code.(type) {
	case int:
		if v != 418 {
			t.Fatalf("Unexpected API code integer: %d; expected 418", v)
		}
	default:
		t.Fatalf("Unexpected API error code type: %T", v)
	}
}

func TestAPIErrorUnmarshalJSONString(t *testing.T) {
	var apiErr APIError
	response := `{"code":"teapot","message":"I'm a teapot","param":"prompt","type":"teapot_error"}`
	err := json.Unmarshal([]byte(response), &apiErr)
	checks.NoError(t, err, "Unexpected Unmarshal API response error")

	switch v := apiErr.Code.(type) {
	case string:
		if v != "teapot" {
			t.Fatalf("Unexpected API code string: %s; expected `teapot`", v)
		}
	default:
		t.Fatalf("Unexpected API error code type: %T", v)
	}
}

func TestAPIErrorUnmarshalJSONNoCode(t *testing.T) {
	// test integer code
	response := `{"message":"I'm a teapot","param":"prompt","type":"teapot_error"}`
	var apiErr APIError
	err := json.Unmarshal([]byte(response), &apiErr)
	checks.NoError(t, err, "Unexpected Unmarshal API response error")

	switch v := apiErr.Code.(type) {
	case nil:
	default:
		t.Fatalf("Unexpected API error code type: %T", v)
	}
}

func TestAPIErrorUnmarshalInvalidData(t *testing.T) {
	apiErr := APIError{}
	data := []byte(`--- {"code":418,"message":"I'm a teapot","param":"prompt","type":"teapot_error"}`)
	err := apiErr.UnmarshalJSON(data)
	checks.HasError(t, err, "Expected error when unmarshaling invalid data")

	if apiErr.Code != nil {
		t.Fatalf("Expected nil code, got %q", apiErr.Code)
	}
	if apiErr.Message != "" {
		t.Fatalf("Expected empty message, got %q", apiErr.Message)
	}
	if apiErr.Param != nil {
		t.Fatalf("Expected nil param, got %q", *apiErr.Param)
	}
	if apiErr.Type != "" {
		t.Fatalf("Expected empty type, got %q", apiErr.Type)
	}
}

func TestAPIErrorUnmarshalJSONInvalidParam(t *testing.T) {
	var apiErr APIError
	response := `{"code":418,"message":"I'm a teapot","param":true,"type":"teapot_error"}`
	err := json.Unmarshal([]byte(response), &apiErr)
	checks.HasError(t, err, "Param should be a string")
}

func TestAPIErrorUnmarshalJSONInvalidType(t *testing.T) {
	var apiErr APIError
	response := `{"code":418,"message":"I'm a teapot","param":"prompt","type":true}`
	err := json.Unmarshal([]byte(response), &apiErr)
	checks.HasError(t, err, "Type should be a string")
}

func TestAPIErrorUnmarshalJSONInvalidMessage(t *testing.T) {
	var apiErr APIError
	response := `{"code":418,"message":false,"param":"prompt","type":"teapot_error"}`
	err := json.Unmarshal([]byte(response), &apiErr)
	checks.HasError(t, err, "Message should be a string")
}

func TestRequestError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/engines", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	_, err := client.ListEngines(context.Background())
	checks.HasError(t, err, "ListEngines did not fail")

	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("Error is not a RequestError: %+v", err)
	}

	if reqErr.HTTPStatusCode != 418 {
		t.Fatalf("Unexpected request error status code: %d", reqErr.HTTPStatusCode)
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
