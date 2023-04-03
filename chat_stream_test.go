package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer server.Close()

	// Client portion of the test
	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = server.URL + "/v1"
	config.HTTPClient.Transport = &tokenRoundTripper{
		test.GetTestToken(),
		http.DefaultTransport,
	}

	client := NewClientWithConfig(config)
	ctx := context.Background()

	request := ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, request)
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer server.Close()

	// Client portion of the test
	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = server.URL + "/v1"
	config.HTTPClient.Transport = &tokenRoundTripper{
		test.GetTestToken(),
		http.DefaultTransport,
	}

	client := NewClientWithConfig(config)
	ctx := context.Background()

	request := ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, request)
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
