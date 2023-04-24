package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"

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

	client := NewClient(test.GetTestToken(), WithCustomBaseURL(ts.URL+"/v1"))
	ctx := context.Background()

	_, err = client.ListModels(ctx)
	checks.NoError(t, err, "ListModels error")
}

// handleModelsEndpoint Handles the models endpoint by the test server.
func handleModelsEndpoint(w http.ResponseWriter, _ *http.Request) {
	resBytes, _ := json.Marshal(ModelsList{})
	fmt.Fprintln(w, string(resBytes))
}
