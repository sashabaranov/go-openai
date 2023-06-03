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

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err = client.ListModels(ctx)
	checks.NoError(t, err, "ListModels error")
}

func TestAzureListModels(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/openai/models", handleModelsEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultAzureConfig(test.GetTestToken(), "https://dummylab.openai.azure.com/")
	config.BaseURL = ts.URL
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err = client.ListModels(ctx)
	checks.NoError(t, err, "ListModels error")
}

// handleModelsEndpoint Handles the models endpoint by the test server.
func handleModelsEndpoint(w http.ResponseWriter, _ *http.Request) {
	resBytes, _ := json.Marshal(ModelsList{})
	fmt.Fprintln(w, string(resBytes))
}

// TestGetModel Tests the retrieve model endpoint of the API using the mocked server.
func TestGetModel(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/models/text-davinci-003", handleGetModelEndpoint)
	// create the test server
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err := client.GetModel(ctx, "text-davinci-003")
	checks.NoError(t, err, "GetModel error")
}

func TestAzureGetModel(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/openai/models/text-davinci-003", handleModelsEndpoint)
	// create the test server
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultAzureConfig(test.GetTestToken(), "https://dummylab.openai.azure.com/")
	config.BaseURL = ts.URL
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err := client.GetModel(ctx, "text-davinci-003")
	checks.NoError(t, err, "GetModel error")
}

// handleModelsEndpoint Handles the models endpoint by the test server.
func handleGetModelEndpoint(w http.ResponseWriter, _ *http.Request) {
	resBytes, _ := json.Marshal(Model{})
	fmt.Fprintln(w, string(resBytes))
}
