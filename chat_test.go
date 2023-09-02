package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
	"github.com/sashabaranov/go-openai/jsonschema"
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
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", handleChatCompletionEndpoint)
	_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	})
	checks.NoError(t, err, "CreateChatCompletion error")
}

// TestChatCompletionsFunctions tests including a function call.
func TestChatCompletionsFunctions(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", handleChatCompletionEndpoint)
	t.Run("bytes", func(t *testing.T) {
		//nolint:lll
		msg := json.RawMessage(`{"properties":{"count":{"type":"integer","description":"total number of words in sentence"},"words":{"items":{"type":"string"},"type":"array","description":"list of words in sentence"}},"type":"object","required":["count","words"]}`)
		_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
			MaxTokens: 5,
			Model:     GPT3Dot5Turbo0613,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
			Functions: []FunctionDefinition{{
				Name:       "test",
				Parameters: &msg,
			}},
		})
		checks.NoError(t, err, "CreateChatCompletion with functions error")
	})
	t.Run("struct", func(t *testing.T) {
		type testMessage struct {
			Count int      `json:"count"`
			Words []string `json:"words"`
		}
		msg := testMessage{
			Count: 2,
			Words: []string{"hello", "world"},
		}
		_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
			MaxTokens: 5,
			Model:     GPT3Dot5Turbo0613,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
			Functions: []FunctionDefinition{{
				Name:       "test",
				Parameters: &msg,
			}},
		})
		checks.NoError(t, err, "CreateChatCompletion with functions error")
	})
	t.Run("JSONSchemaDefinition", func(t *testing.T) {
		_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
			MaxTokens: 5,
			Model:     GPT3Dot5Turbo0613,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
			Functions: []FunctionDefinition{{
				Name: "test",
				Parameters: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"count": {
							Type:        jsonschema.Number,
							Description: "total number of words in sentence",
						},
						"words": {
							Type:        jsonschema.Array,
							Description: "list of words in sentence",
							Items: &jsonschema.Definition{
								Type: jsonschema.String,
							},
						},
						"enumTest": {
							Type: jsonschema.String,
							Enum: []string{"hello", "world"},
						},
					},
				},
			}},
		})
		checks.NoError(t, err, "CreateChatCompletion with functions error")
	})
	t.Run("JSONSchemaDefinitionWithFunctionDefine", func(t *testing.T) {
		// this is a compatibility check
		_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
			MaxTokens: 5,
			Model:     GPT3Dot5Turbo0613,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
			Functions: []FunctionDefine{{
				Name: "test",
				Parameters: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"count": {
							Type:        jsonschema.Number,
							Description: "total number of words in sentence",
						},
						"words": {
							Type:        jsonschema.Array,
							Description: "list of words in sentence",
							Items: &jsonschema.Definition{
								Type: jsonschema.String,
							},
						},
						"enumTest": {
							Type: jsonschema.String,
							Enum: []string{"hello", "world"},
						},
					},
				},
			}},
		})
		checks.NoError(t, err, "CreateChatCompletion with functions error")
	})
}

func TestAzureChatCompletions(t *testing.T) {
	client, server, teardown := setupAzureTestServer()
	defer teardown()
	server.RegisterHandler("/openai/deployments/*", handleChatCompletionEndpoint)

	_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	})
	checks.NoError(t, err, "CreateAzureChatCompletion error")
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
	n := completionReq.N
	if n == 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		// if there are functions, include them
		if len(completionReq.Functions) > 0 {
			var fcb []byte
			b := completionReq.Functions[0].Parameters
			fcb, err = json.Marshal(b)
			if err != nil {
				http.Error(w, "could not marshal function parameters", http.StatusInternalServerError)
				return
			}

			res.Choices = append(res.Choices, ChatCompletionChoice{
				Message: ChatCompletionMessage{
					Role: ChatMessageRoleFunction,
					// this is valid json so it should be fine
					FunctionCall: &FunctionCall{
						Name:      completionReq.Functions[0].Name,
						Arguments: string(fcb),
					},
				},
				Index: i,
			})
			continue
		}
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
	inputTokens := numTokens(completionReq.Messages[0].Content) * n
	completionTokens := completionReq.MaxTokens * n
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

func TestFinishReason(t *testing.T) {
	c := &ChatCompletionChoice{
		FinishReason: FinishReasonNull,
	}
	resBytes, _ := json.Marshal(c)
	if !strings.Contains(string(resBytes), `"finish_reason":null`) {
		t.Error("null should not be quoted")
	}

	c.FinishReason = ""

	resBytes, _ = json.Marshal(c)
	if !strings.Contains(string(resBytes), `"finish_reason":null`) {
		t.Error("null should not be quoted")
	}

	otherReasons := []FinishReason{
		FinishReasonStop,
		FinishReasonLength,
		FinishReasonFunctionCall,
		FinishReasonContentFilter,
	}
	for _, r := range otherReasons {
		c.FinishReason = r
		resBytes, _ = json.Marshal(c)
		if !strings.Contains(string(resBytes), fmt.Sprintf(`"finish_reason":"%s"`, r)) {
			t.Errorf("%s should be quoted", r)
		}
	}
}
