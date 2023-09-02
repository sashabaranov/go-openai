package openai_test

import (
	"os"
	"time"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TestListModels Tests the list models endpoint of the API using the mocked server.
func TestListModels(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/models", handleListModelsEndpoint)
	_, err := client.ListModels(context.Background())
	checks.NoError(t, err, "ListModels error")
}

func TestAzureListModels(t *testing.T) {
	client, server, teardown := setupAzureTestServer()
	defer teardown()
	server.RegisterHandler("/openai/models", handleListModelsEndpoint)
	_, err := client.ListModels(context.Background())
	checks.NoError(t, err, "ListModels error")
}

// handleListModelsEndpoint Handles the list models endpoint by the test server.
func handleListModelsEndpoint(w http.ResponseWriter, _ *http.Request) {
	resBytes, _ := json.Marshal(ModelsList{})
	fmt.Fprintln(w, string(resBytes))
}

// TestGetModel Tests the retrieve model endpoint of the API using the mocked server.
func TestGetModel(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/models/text-davinci-003", handleGetModelEndpoint)
	_, err := client.GetModel(context.Background(), "text-davinci-003")
	checks.NoError(t, err, "GetModel error")
}

func TestAzureGetModel(t *testing.T) {
	client, server, teardown := setupAzureTestServer()
	defer teardown()
	server.RegisterHandler("/openai/models/text-davinci-003", handleGetModelEndpoint)
	_, err := client.GetModel(context.Background(), "text-davinci-003")
	checks.NoError(t, err, "GetModel error")
}

// handleGetModelsEndpoint Handles the get model endpoint by the test server.
func handleGetModelEndpoint(w http.ResponseWriter, _ *http.Request) {
	resBytes, _ := json.Marshal(Model{})
	fmt.Fprintln(w, string(resBytes))
}

func TestGetModelReturnTimeoutError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/models/text-davinci-003", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Nanosecond)
	})
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	defer cancel()

	_, err := client.GetModel(ctx, "text-davinci-003")
	if err == nil {
		t.Fatal("Did not return error")
	}
	if !os.IsTimeout(err) {
		t.Fatal("Did not return timeout error")
	}
}
