package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"

	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCompletionStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}
		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data := `{"id":"1","object":"completion","created":1598069254,"model":"text-davinci-002","choices":[{"text":"response1","finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data = `{"id":"2","object":"completion","created":1598069255,"model":"text-davinci-002","choices":[{"text":"response2","finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: done\n")...)
		dataBytes = append(dataBytes, []byte("data: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		if err != nil {
			t.Errorf("Write error: %s", err)
		}
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

	request := CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     "text-davinci-002",
		MaxTokens: 10,
		Stream:    true,
	}

	stream, err := client.CreateCompletionStream(ctx, request)
	if err != nil {
		t.Errorf("CreateCompletionStream returned error: %v", err)
	}
	defer stream.Close()

	expectedResponses := []CompletionResponse{
		{
			ID:      "1",
			Object:  "completion",
			Created: 1598069254,
			Model:   "text-davinci-002",
			Choices: []CompletionChoice{{Text: "response1", FinishReason: "max_tokens"}},
		},
		{
			ID:      "2",
			Object:  "completion",
			Created: 1598069255,
			Model:   "text-davinci-002",
			Choices: []CompletionChoice{{Text: "response2", FinishReason: "max_tokens"}},
		},
	}

	for ix, expectedResponse := range expectedResponses {
		receivedResponse, streamErr := stream.Recv()
		if streamErr != nil {
			t.Errorf("stream.Recv() failed: %v", streamErr)
		}
		if !compareResponses(expectedResponse, receivedResponse) {
			t.Errorf("Stream response %v is %v, expected %v", ix, receivedResponse, expectedResponse)
		}
	}

	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF in the end: %v", streamErr)
	}
}

// A "tokenRoundTripper" is a struct that implements the RoundTripper
// interface, specifically to handle the authentication token by adding a token
// to the request header. We need this because the API requires that each
// request include a valid API token in the headers for authentication and
// authorization.
type tokenRoundTripper struct {
	token    string
	fallback http.RoundTripper
}

// RoundTrip takes an *http.Request as input and returns an
// *http.Response and an error.
//
// It is expected to use the provided request to create a connection to an HTTP
// server and return the response, or an error if one occurred. The returned
// Response should have its Body closed. If the RoundTrip method returns an
// error, the Client's Get, Head, Post, and PostForm methods return the same
// error.
func (t *tokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.fallback.RoundTrip(req)
}

// Helper funcs.
func compareResponses(r1, r2 CompletionResponse) bool {
	if r1.ID != r2.ID || r1.Object != r2.Object || r1.Created != r2.Created || r1.Model != r2.Model {
		return false
	}
	if len(r1.Choices) != len(r2.Choices) {
		return false
	}
	for i := range r1.Choices {
		if !compareResponseChoices(r1.Choices[i], r2.Choices[i]) {
			return false
		}
	}
	return true
}

func compareResponseChoices(c1, c2 CompletionChoice) bool {
	if c1.Text != c2.Text || c1.FinishReason != c2.FinishReason {
		return false
	}
	return true
}
