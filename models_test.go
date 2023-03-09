package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TestListModels Tests the models endpoint of the API using the mocked server.
func TestListModels(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/models", handleModelsEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err = client.ListModels(ctx)
	if err != nil {
		t.Fatalf("ListModels error: %v", err)
	}
}

// handleModelsEndpoint Handles the models endpoint by the test server.
func handleModelsEndpoint(w http.ResponseWriter, r *http.Request) {
	resBytes, _ := json.Marshal(ModelsList{})
	fmt.Fprintln(w, string(resBytes))
}
