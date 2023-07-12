package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestChatCompletionsStreamWrongModel(t *testing.T) {
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
	_, err := client.CreateChatCompletionStream(ctx, req)
	if !errors.Is(err, ErrChatCompletionInvalidModel) {
		t.Fatalf("CreateChatCompletion should return ErrChatCompletionInvalidModel, but returned: %v", err)
	}
}

func TestCreateChatCompletionStream(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}
		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data := `{"id":"1","object":"completion","created":1598069254,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"response1"},"finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data = `{"id":"2","object":"completion","created":1598069255,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"response2"},"finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: done\n")...)
		dataBytes = append(dataBytes, []byte("data: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	expectedResponses := []ChatCompletionStreamResponse{
		{
			ID:      "1",
			Object:  "completion",
			Created: 1598069254,
			Model:   GPT3Dot5Turbo,
			Choices: []ChatCompletionStreamChoice{
				{
					Delta: ChatCompletionStreamChoiceDelta{
						Content: "response1",
					},
					FinishReason: "max_tokens",
				},
			},
		},
		{
			ID:      "2",
			Object:  "completion",
			Created: 1598069255,
			Model:   GPT3Dot5Turbo,
			Choices: []ChatCompletionStreamChoice{
				{
					Delta: ChatCompletionStreamChoiceDelta{
						Content: "response2",
					},
					FinishReason: "max_tokens",
				},
			},
		},
	}

	for ix, expectedResponse := range expectedResponses {
		b, _ := json.Marshal(expectedResponse)
		t.Logf("%d: %s", ix, string(b))

		receivedResponse, streamErr := stream.Recv()
		checks.NoError(t, streamErr, "stream.Recv() failed")
		if !compareChatResponses(expectedResponse, receivedResponse) {
			t.Errorf("Stream response %v is %v, expected %v", ix, receivedResponse, expectedResponse)
		}
	}

	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF in the end: %v", streamErr)
	}

	_, streamErr = stream.Recv()

	checks.ErrorIs(t, streamErr, io.EOF, "stream.Recv() did not return EOF when the stream is finished")
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF when the stream is finished: %v", streamErr)
	}
}

func TestCreateChatCompletionStreamError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}
		dataStr := []string{
			`{`,
			`"error": {`,
			`"message": "Incorrect API key provided: sk-***************************************",`,
			`"type": "invalid_request_error",`,
			`"param": null,`,
			`"code": "invalid_api_key"`,
			`}`,
			`}`,
		}
		for _, str := range dataStr {
			dataBytes = append(dataBytes, []byte(str+"\n")...)
		}

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	_, streamErr := stream.Recv()
	checks.HasError(t, streamErr, "stream.Recv() did not return error")

	var apiErr *APIError
	if !errors.As(streamErr, &apiErr) {
		t.Errorf("stream.Recv() did not return APIError")
	}
	t.Logf("%+v\n", apiErr)
}

func TestCreateChatCompletionStreamErrorWithDataPrefix(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		//nolint:lll
		dataBytes := []byte(`data: {"error":{"message":"The server had an error while processing your request. Sorry about that!", "type":"server_ error", "param":null,"code":null}}`)
		dataBytes = append(dataBytes, []byte("\n\ndata: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	_, streamErr := stream.Recv()
	checks.HasError(t, streamErr, "stream.Recv() did not return error")

	var apiErr *APIError
	if !errors.As(streamErr, &apiErr) {
		t.Errorf("stream.Recv() did not return APIError")
	}
	t.Logf("%+v\n", apiErr)
}

func TestCreateChatCompletionStreamRateLimitError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)

		// Send test responses
		dataBytes := []byte(`{"error":{` +
			`"message": "You are sending requests too quickly.",` +
			`"type":"rate_limit_reached",` +
			`"param":null,` +
			`"code":"rate_limit_reached"}}`)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})
	_, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("TestCreateChatCompletionStreamRateLimitError did not return APIError")
	}
	t.Logf("%+v\n", apiErr)
}

func TestAzureCreateChatCompletionStreamRateLimitError(t *testing.T) {
	wantCode := "429"
	wantMessage := "Requests to the Creates a completion for the chat message Operation under Azure OpenAI API " +
		"version 2023-03-15-preview have exceeded token rate limit of your current OpenAI S0 pricing tier. " +
		"Please retry after 20 seconds. " +
		"Please go here: https://aka.ms/oai/quotaincrease if you would like to further increase the default rate limit."

	client, server, teardown := setupAzureTestServer()
	defer teardown()
	server.RegisterHandler("/openai/deployments/gpt-35-turbo/chat/completions",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			// Send test responses
			dataBytes := []byte(`{"error": { "code": "` + wantCode + `", "message": "` + wantMessage + `"}}`)
			_, err := w.Write(dataBytes)

			checks.NoError(t, err, "Write error")
		})

	apiErr := &APIError{}
	_, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	if !errors.As(err, &apiErr) {
		t.Errorf("Did not return APIError: %+v\n", apiErr)
		return
	}
	if apiErr.HTTPStatusCode != http.StatusTooManyRequests {
		t.Errorf("Did not return HTTPStatusCode got = %d, want = %d\n", apiErr.HTTPStatusCode, http.StatusTooManyRequests)
		return
	}
	code, ok := apiErr.Code.(string)
	if !ok || code != wantCode {
		t.Errorf("Did not return Code. got = %v, want = %s\n", apiErr.Code, wantCode)
		return
	}
	if apiErr.Message != wantMessage {
		t.Errorf("Did not return Message. got = %s, want = %s\n", apiErr.Message, wantMessage)
		return
	}
}

// Helper funcs.
func compareChatResponses(r1, r2 ChatCompletionStreamResponse) bool {
	if r1.ID != r2.ID || r1.Object != r2.Object || r1.Created != r2.Created || r1.Model != r2.Model {
		return false
	}
	if len(r1.Choices) != len(r2.Choices) {
		return false
	}
	for i := range r1.Choices {
		if !compareChatStreamResponseChoices(r1.Choices[i], r2.Choices[i]) {
			return false
		}
	}
	return true
}

func compareChatStreamResponseChoices(c1, c2 ChatCompletionStreamChoice) bool {
	if c1.Index != c2.Index {
		return false
	}
	if c1.Delta.Content != c2.Delta.Content {
		return false
	}
	if c1.FinishReason != c2.FinishReason {
		return false
	}
	return true
}
