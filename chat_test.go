package openai_test

import (
	"errors"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestChatCompletionsWrongModel(t *testing.T) {
	config := DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	req := ChatCompletionRequest{
		MaxTokens: 5,
		Model:     "ada",
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	}
	_, err := client.CreateChatCompletion(ctx, req)
	msg := fmt.Sprintf("CreateChatCompletion should return wrong model error, returned: %s", err)
	checks.ErrorIs(t, err, ErrChatCompletionInvalidModel, msg)
}

func TestChatCompletionsWithStream(t *testing.T) {
	config := DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	req := ChatCompletionRequest{
		Stream: true,
	}
	_, err := client.CreateChatCompletion(ctx, req)
	checks.ErrorIs(t, err, ErrChatCompletionStreamNotSupported, "unexpected error")
}

// TestCompletions Tests the completions endpoint of the API using the mocked server.
func TestChatCompletions(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/chat/completions", handleChatCompletionEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	req := ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	}
	_, err = client.CreateChatCompletion(ctx, req)
	checks.NoError(t, err, "CreateChatCompletion error")
}

func TestChatCompletionsRateLimit(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/chat/completions", handleChatCompletionEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.EnableRateLimiter = true
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	req := ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	}
	_, err = client.CreateChatCompletion(ctx, req)
	checks.NoError(t, err, "CreateChatCompletion error")
}

func TestChatCompletionsRateLimitErr(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/chat/completions", handleChatCompletionEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.EnableRateLimiter = true
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx, cancel := context.WithCancel(context.Background())
	req := ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	}
	cancel()
	_, err = client.CreateChatCompletion(ctx, req)
	checks.HasError(t, err, "CreateChatCompletion error")
}

// handleChatCompletionEndpoint Handles the ChatGPT completion endpoint by the test server.
func handleChatCompletionEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	var completionReq ChatCompletionRequest
	if completionReq, err = getChatCompletionBody(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}
	res := ChatCompletionResponse{
		ID:      strconv.Itoa(int(time.Now().Unix())),
		Object:  "test-object",
		Created: time.Now().Unix(),
		// would be nice to validate Model during testing, but
		// this may not be possible with how much upkeep
		// would be required / wouldn't make much sense
		Model: completionReq.Model,
	}
	// create completions
	for i := 0; i < completionReq.N; i++ {
		// generate a random string of length completionReq.Length
		completionStr := strings.Repeat("a", completionReq.MaxTokens)

		res.Choices = append(res.Choices, ChatCompletionChoice{
			Message: ChatCompletionMessage{
				Role:    ChatMessageRoleAssistant,
				Content: completionStr,
			},
			Index: i,
		})
	}
	inputTokens := numTokens(completionReq.Messages[0].Content) * completionReq.N
	completionTokens := completionReq.MaxTokens * completionReq.N
	res.Usage = Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      inputTokens + completionTokens,
	}
	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}

// getChatCompletionBody Returns the body of the request to create a completion.
func getChatCompletionBody(r *http.Request) (ChatCompletionRequest, error) {
	completion := ChatCompletionRequest{}
	// read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return ChatCompletionRequest{}, err
	}
	err = json.Unmarshal(reqBody, &completion)
	if err != nil {
		return ChatCompletionRequest{}, err
	}
	return completion, nil
}

func TestChatCompletionRequest_Tokens(t *testing.T) {
	testcases := []struct {
		name       string
		model      string
		messages   []ChatCompletionMessage
		wantErr    error
		wantTokens int
	}{
		{
			name:    "test unknown model",
			model:   "unknown",
			wantErr: errors.New("failed to tokenize prompt: model not supported: model not supported"),
		},
		{
			name:       "test1",
			model:      GPT3Dot5Turbo,
			messages:   []ChatCompletionMessage{{Content: "Hello!"}},
			wantTokens: 2,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(tt *testing.T) {
			req := ChatCompletionRequest{
				Model:    testcase.model,
				Messages: testcase.messages,
			}
			tokens, err := req.Tokens()
			if err != nil && testcase.wantErr == nil {
				tt.Fatalf("Tokens() returned unexpected error: %v", err)
			}

			if err != nil && testcase.wantErr != nil && err.Error() != testcase.wantErr.Error() {
				tt.Fatalf("Tokens() returned unexpected error: %v, want: %v", err, testcase.wantErr)
			}

			if tokens != testcase.wantTokens {
				tt.Fatalf("Tokens() returned unexpected number of tokens: %d, want: %d", tokens, testcase.wantTokens)
			}
		})
	}
}
