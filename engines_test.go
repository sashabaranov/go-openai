package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

// TestGetEngine Tests the retrieve engine endpoint of the API using the mocked server.
func TestGetEngine(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/engines/text-davinci-003", func(w http.ResponseWriter, r *http.Request) {
		resBytes, _ := json.Marshal(Engine{})
		fmt.Fprintln(w, string(resBytes))
	})
	_, err := client.GetEngine(context.Background(), "text-davinci-003")
	checks.NoError(t, err, "GetEngine error")
}

// TestListEngines Tests the list engines endpoint of the API using the mocked server.
func TestListEngines(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/engines", func(w http.ResponseWriter, r *http.Request) {
		resBytes, _ := json.Marshal(EnginesList{})
		fmt.Fprintln(w, string(resBytes))
	})
	_, err := client.ListEngines(context.Background())
	checks.NoError(t, err, "ListEngines error")
}

func TestListEnginesReturnError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/engines", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	_, err := client.ListEngines(context.Background())
	checks.HasError(t, err, "ListEngines did not fail")
}
