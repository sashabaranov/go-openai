package openai_test

import (
	. "github.com/sashabaranov/go-openai"
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
	t.Run("ParametersRaw", func(t *testing.T) {
		_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
			MaxTokens: 5,
			Model:     GPT3Dot5Turbo,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
			Functions: []*FunctionDefine{{
				Name: "test",
				//nolint:lll
				ParametersRaw: json.RawMessage(`{"properties":{"count":{"type":"integer","description":"total number of words in sentence"},"words":{"items":{"type":"string"},"type":"array","description":"list of words in sentence"}},"type":"object","required":["count","words"]}`),
			}},
		})
		checks.NoError(t, err, "CreateChatCompletion with functions and parametersRaw error")
	})
	t.Run("Parameters", func(t *testing.T) {
		_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
			MaxTokens: 5,
			Model:     GPT3Dot5Turbo,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
			Functions: []*FunctionDefine{{
				Name: "test",
				Parameters: &FunctionParams{
					Properties: map[string]*JSONSchemaDefine{
						"count": {
							Type:        "integer",
							Description: "total number of words in sentence",
						},
						"words": {
							Type:        "array",
							Description: "list of words in sentence",
							Items: &JSONSchemaDefine{
								Type: JSONSchemaTypeString,
							},
						},
					},
					Required: []string{"count", "words"},
					Type:     "object",
				},
			}},
		})
		checks.NoError(t, err, "CreateChatCompletion with functions and parametersRaw error")
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
			var fc map[string]interface{}
			b := completionReq.Functions[0].ParametersRaw
			if completionReq.Functions[0].Parameters != nil {
				// marshal this to json
				b, err = json.Marshal(completionReq.Functions[0].Parameters)
				if err != nil {
					http.Error(w, "could not marshal function parameters", http.StatusInternalServerError)
					return
				}
			}

			if err = json.Unmarshal(b, &fc); err != nil {
				http.Error(w, "could not unmarshal function parameters", http.StatusInternalServerError)
				return
			}
			res.Choices = append(res.Choices, ChatCompletionChoice{
				Message: ChatCompletionMessage{
					Role: ChatMessageRoleFunction,
					// this is valid json so it should be fine
					FunctionCall: &FunctionCall{
						Name:      completionReq.Functions[0].Name,
						Arguments: string(completionReq.Functions[0].ParametersRaw),
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

func TestMarshalJSON(t *testing.T) {
	t.Run("ParametersRaw is not nil", func(t *testing.T) {
		funcDefine := FunctionDefine{Name: "testFunc", ParametersRaw: json.RawMessage(`{"name":"test"}`)}

		expected := `{"name":"testFunc","parameters":{"name":"test"}}`
		b, err := funcDefine.MarshalJSON()
		checks.NoError(t, err)

		if string(b) != expected {
			t.Errorf("Got %v, expected %v", string(b), expected)
		}
	})

	t.Run("ParametersRaw is nil, Parameters is not nil", func(t *testing.T) {
		params := &FunctionParams{
			Type:       JSONSchemaTypeObject,
			Properties: map[string]*JSONSchemaDefine{"name": {Type: JSONSchemaTypeString}},
		}
		funcDefine := FunctionDefine{Name: "testFunc", Parameters: params}

		expected := `{"name":"testFunc","parameters":{"type":"object","properties":{"name":{"type":"string"}}}}`
		b, err := funcDefine.MarshalJSON()
		checks.NoError(t, err)

		if string(b) != expected {
			t.Errorf("Got %v, expected %v", string(b), expected)
		}
	})

	t.Run("ParametersRaw is not nil, Parameters is not nil", func(t *testing.T) {
		params := &FunctionParams{
			Type:       JSONSchemaTypeObject,
			Properties: map[string]*JSONSchemaDefine{"name": {Type: JSONSchemaTypeString}},
		}
		funcDefine := FunctionDefine{Name: "testFunc", ParametersRaw: json.RawMessage(`{"name":"test"}`), Parameters: params}

		expected := `{"name":"testFunc","parameters":{"name":"test"}}`
		b, err := funcDefine.MarshalJSON()
		checks.NoError(t, err)

		if string(b) != expected {
			t.Errorf("Got %v, expected %v", string(b), expected)
		}
	})
}

func TestUnmarshalJSON(t *testing.T) {
	t.Run("ParametersRaw is valid", func(t *testing.T) {
		data := []byte(`{"name":"testFunc","parameters":{"type":"object","properties":{"name":{"type":"string"}}}}`)

		var funcDefine FunctionDefine
		err := funcDefine.UnmarshalJSON(data)
		checks.NoError(t, err)

		if funcDefine.Name != "testFunc" {
			t.Errorf("Got %v, expected testFunc", funcDefine.Name)
		}

		if funcDefine.Parameters.Type != JSONSchemaTypeObject {
			t.Errorf("Got %v, expected object", funcDefine.Parameters.Type)
		}
	})

	t.Run("ParametersRaw is invalid", func(t *testing.T) {
		data := []byte(`{"name":"testFunc","parameters":"invalid json"}`)

		var funcDefine FunctionDefine
		err := funcDefine.UnmarshalJSON(data)
		checks.HasError(t, err, "invalid character")
	})
}
