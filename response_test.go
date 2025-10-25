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

func TestCreateResponse(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/responses", handleResponseEndpoint)
	_, err := client.CreateResponse(context.Background(), openai.CreateResponseRequest{
		Model: "gpt-4o",
		Input: "What's the latest news about AI?",
		Tools: []openai.Tool{
			{
				Type: "web_search",
			},
		},
	})
	checks.NoError(t, err, "CreateResponse error")
}

func handleResponseEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var responseReq openai.CreateResponseRequest
	if err = json.NewDecoder(r.Body).Decode(&responseReq); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}

	res := openai.CreateResponseResponse{
		ID:      "resp_" + strconv.Itoa(int(time.Now().Unix())),
		Created: time.Now().Unix(),
		Model:   responseReq.Model,
		Output:  []any{},
	}

	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}
