package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestCompletionsWrongModel(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := openai.NewClientWithConfig(config)

	_, err := client.CreateCompletion(
		context.Background(),
		openai.CompletionRequest{
			MaxTokens: 5,
			Model:     openai.GPT3Dot5Turbo,
		},
	)
	if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
		t.Fatalf("CreateCompletion should return ErrCompletionUnsupportedModel, but returned: %v", err)
	}
}

// TestCompletionsWrongModelO3 Tests the completions endpoint with O3 model which is not supported.
func TestCompletionsWrongModelO3(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := openai.NewClientWithConfig(config)

	_, err := client.CreateCompletion(
		context.Background(),
		openai.CompletionRequest{
			MaxTokens: 5,
			Model:     openai.O3,
		},
	)
	if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
		t.Fatalf("CreateCompletion should return ErrCompletionUnsupportedModel for O3, but returned: %v", err)
	}
}

// TestCompletionsWrongModelO4Mini Tests the completions endpoint with O4Mini model which is not supported.
func TestCompletionsWrongModelO4Mini(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := openai.NewClientWithConfig(config)

	_, err := client.CreateCompletion(
		context.Background(),
		openai.CompletionRequest{
			MaxTokens: 5,
			Model:     openai.O4Mini,
		},
	)
	if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
		t.Fatalf("CreateCompletion should return ErrCompletionUnsupportedModel for O4Mini, but returned: %v", err)
	}
}

func TestCompletionWithStream(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	client := openai.NewClientWithConfig(config)

	ctx := context.Background()
	req := openai.CompletionRequest{Stream: true}
	_, err := client.CreateCompletion(ctx, req)
	if !errors.Is(err, openai.ErrCompletionStreamNotSupported) {
		t.Fatalf("CreateCompletion didn't return ErrCompletionStreamNotSupported")
	}
}

// TestCompletions Tests the completions endpoint of the API using the mocked server.
func TestCompletions(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", handleCompletionEndpoint)
	req := openai.CompletionRequest{
		MaxTokens: 5,
		Model:     "ada",
		Prompt:    "Lorem ipsum",
	}
	_, err := client.CreateCompletion(context.Background(), req)
	checks.NoError(t, err, "CreateCompletion error")
}

// TestMultiplePromptsCompletionsWrong Tests the completions endpoint of the API using the mocked server
// where the completions requests has a list of prompts with wrong type.
func TestMultiplePromptsCompletionsWrong(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", handleCompletionEndpoint)
	req := openai.CompletionRequest{
		MaxTokens: 5,
		Model:     "ada",
		Prompt:    []interface{}{"Lorem ipsum", 9},
	}
	_, err := client.CreateCompletion(context.Background(), req)
	if !errors.Is(err, openai.ErrCompletionRequestPromptTypeNotSupported) {
		t.Fatalf("CreateCompletion should return ErrCompletionRequestPromptTypeNotSupported, but returned: %v", err)
	}
}

// TestMultiplePromptsCompletions Tests the completions endpoint of the API using the mocked server
// where the completions requests has a list of prompts.
func TestMultiplePromptsCompletions(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", handleCompletionEndpoint)
	req := openai.CompletionRequest{
		MaxTokens: 5,
		Model:     "ada",
		Prompt:    []interface{}{"Lorem ipsum", "Lorem ipsum"},
	}
	_, err := client.CreateCompletion(context.Background(), req)
	checks.NoError(t, err, "CreateCompletion error")
}

// handleCompletionEndpoint Handles the completion endpoint by the test server.
func handleCompletionEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	var completionReq openai.CompletionRequest
	if completionReq, err = getCompletionBody(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}
	res := openai.CompletionResponse{
		ID:      strconv.Itoa(int(time.Now().Unix())),
		Object:  "test-object",
		Created: time.Now().Unix(),
		// would be nice to validate Model during testing, but
		// this may not be possible with how much upkeep
		// would be required / wouldn't make much sense
		Model: completionReq.Model,
	}
	// create completions
	n := completionReq.N
	if n == 0 {
		n = 1
	}
	// Handle different types of prompts: single string or list of strings
	prompts := []string{}
	switch v := completionReq.Prompt.(type) {
	case string:
		prompts = append(prompts, v)
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				prompts = append(prompts, str)
			}
		}
	default:
		http.Error(w, "Invalid prompt type", http.StatusBadRequest)
		return
	}

	for i := 0; i < n; i++ {
		for _, prompt := range prompts {
			// Generate a random string of length completionReq.MaxTokens
			completionStr := strings.Repeat("a", completionReq.MaxTokens)
			if completionReq.Echo {
				completionStr = prompt + completionStr
			}

			res.Choices = append(res.Choices, openai.CompletionChoice{
				Text:  completionStr,
				Index: len(res.Choices),
			})
		}
	}

	inputTokens := 0
	for _, prompt := range prompts {
		inputTokens += numTokens(prompt)
	}
	inputTokens *= n
	completionTokens := completionReq.MaxTokens * len(prompts) * n
	res.Usage = openai.Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      inputTokens + completionTokens,
	}

	// Serialize the response and send it back
	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}

// getCompletionBody Returns the body of the request to create a completion.
func getCompletionBody(r *http.Request) (openai.CompletionRequest, error) {
	completion := openai.CompletionRequest{}
	// read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return openai.CompletionRequest{}, err
	}
	err = json.Unmarshal(reqBody, &completion)
	if err != nil {
		return openai.CompletionRequest{}, err
	}
	return completion, nil
}

// TestCompletionWithO1Model Tests that O1 model is not supported for completion endpoint.
func TestCompletionWithO1Model(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := openai.NewClientWithConfig(config)

	_, err := client.CreateCompletion(
		context.Background(),
		openai.CompletionRequest{
			MaxTokens: 5,
			Model:     openai.O1,
		},
	)
	if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
		t.Fatalf("CreateCompletion should return ErrCompletionUnsupportedModel for O1 model, but returned: %v", err)
	}
}

// TestCompletionWithGPT4DotModels Tests that newer GPT4 models are not supported for completion endpoint.
func TestCompletionWithGPT4DotModels(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := openai.NewClientWithConfig(config)

	models := []string{
		openai.GPT4Dot1,
		openai.GPT4Dot120250414,
		openai.GPT4Dot1Mini,
		openai.GPT4Dot1Mini20250414,
		openai.GPT4Dot1Nano,
		openai.GPT4Dot1Nano20250414,
		openai.GPT4Dot5Preview,
		openai.GPT4Dot5Preview20250227,
	}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			_, err := client.CreateCompletion(
				context.Background(),
				openai.CompletionRequest{
					MaxTokens: 5,
					Model:     model,
				},
			)
			if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
				t.Fatalf("CreateCompletion should return ErrCompletionUnsupportedModel for %s model, but returned: %v", model, err)
			}
		})
	}
}

// TestCompletionWithGPT4oModels Tests that GPT4o models are not supported for completion endpoint.
func TestCompletionWithGPT4oModels(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := openai.NewClientWithConfig(config)

	models := []string{
		openai.GPT4o,
		openai.GPT4o20240513,
		openai.GPT4o20240806,
		openai.GPT4o20241120,
		openai.GPT4oLatest,
		openai.GPT4oMini,
		openai.GPT4oMini20240718,
	}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			_, err := client.CreateCompletion(
				context.Background(),
				openai.CompletionRequest{
					MaxTokens: 5,
					Model:     model,
				},
			)
			if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
				t.Fatalf("CreateCompletion should return ErrCompletionUnsupportedModel for %s model, but returned: %v", model, err)
			}
		})
	}
}
