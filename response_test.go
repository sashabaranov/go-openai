package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestCreateResponsesAPI(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/responses", handleResponseEndpoint)
	_, err := client.CreateResponsesAPI(context.Background(), openai.ResponsesAPIRequest{
		Model: openai.GPT4o,
		Input: "What's the latest news about AI?",
		Tools: []openai.Tool{
			{
				Type: openai.ToolTypeWebSearch,
			},
		},
	})
	checks.NoError(t, err, "CreateResponsesAPI error")
}

func TestCreateResponsesAPIError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	// Test newRequest error: Invalid body
	_, err := client.CreateResponsesAPI(context.Background(), openai.ResponsesAPIRequest{
		Model: openai.GPT4o,
		Input: make(chan int), // Invalid type for JSON marshalling
	})
	checks.HasError(t, err, "CreateResponsesAPI error expected (newRequest)")

	// Test sendRequest error: Server error
	server.RegisterHandler("/v1/responses", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	_, err = client.CreateResponsesAPI(context.Background(), openai.ResponsesAPIRequest{
		Model: openai.GPT4o,
		Input: "Hello",
	})
	checks.HasError(t, err, "CreateResponsesAPI error expected (sendRequest)")
}

func handleResponseEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var responseReq openai.ResponsesAPIRequest
	if err = json.NewDecoder(r.Body).Decode(&responseReq); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}

	res := openai.ResponsesAPIResponse{
		ID:      "resp_" + strconv.Itoa(int(time.Now().Unix())),
		Created: time.Now().Unix(),
		Model:   responseReq.Model,
		Output:  []any{},
	}

	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}
