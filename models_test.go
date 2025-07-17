package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/lookfirst-io/go-openai"
	"github.com/lookfirst-io/go-openai/internal/test/checks"
)

const testFineTuneModelID = "fine-tune-model-id"

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
	resBytes, _ := json.Marshal(openai.ModelsList{})
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

// TestGetModelO3 Tests the retrieve O3 model endpoint of the API using the mocked server.
func TestGetModelO3(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/models/o3", handleGetModelEndpoint)
	_, err := client.GetModel(context.Background(), "o3")
	checks.NoError(t, err, "GetModel error for O3")
}

// TestGetModelO4Mini Tests the retrieve O4Mini model endpoint of the API using the mocked server.
func TestGetModelO4Mini(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/models/o4-mini", handleGetModelEndpoint)
	_, err := client.GetModel(context.Background(), "o4-mini")
	checks.NoError(t, err, "GetModel error for O4Mini")
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
	resBytes, _ := json.Marshal(openai.Model{})
	fmt.Fprintln(w, string(resBytes))
}

func TestGetModelReturnTimeoutError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/models/text-davinci-003", func(http.ResponseWriter, *http.Request) {
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

func TestDeleteFineTuneModel(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/models/"+testFineTuneModelID, handleDeleteFineTuneModelEndpoint)
	_, err := client.DeleteFineTuneModel(context.Background(), testFineTuneModelID)
	checks.NoError(t, err, "DeleteFineTuneModel error")
}

func handleDeleteFineTuneModelEndpoint(w http.ResponseWriter, _ *http.Request) {
	resBytes, _ := json.Marshal(openai.FineTuneModelDeleteResponse{})
	fmt.Fprintln(w, string(resBytes))
}
