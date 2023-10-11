package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

// Helper function to test the ListEngines endpoint.
func RandomEngine() Engine {
	return Engine{
		ID:     test.RandomString(),
		Object: test.RandomString(),
		Owner:  test.RandomString(),
		Ready:  test.RandomBool(),
	}
}

// TestGetEngine Tests the retrieve engine endpoint of the API using the mocked server.
func TestGetEngine(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	expectedEngine := RandomEngine() // move outside of handler per code review comment
	server.RegisterHandler("/v1/engines/text-davinci-003", func(w http.ResponseWriter, r *http.Request) {
		resBytes, _ := json.Marshal(expectedEngine)
		fmt.Fprintln(w, string(resBytes))
	})
	actualEngine, err := client.GetEngine(context.Background(), "text-davinci-003")
	checks.NoError(t, err, "GetEngine error")

	// Compare the two using only one field per code review comment
	if actualEngine.ID != expectedEngine.ID {
		t.Errorf("Engine ID mismatch: got %s, expected %s", actualEngine.ID, expectedEngine.ID)
	}
}

// TestListEngines Tests the list engines endpoint of the API using the mocked server.
func TestListEngines(t *testing.T) {
	test.MaybeSeedRNG() // see docstring at internal/test/random.go
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/engines", func(w http.ResponseWriter, r *http.Request) {
		engines := make([]Engine, test.RandomInt(5))
		for i := range engines {
			engines[i] = RandomEngine()
		}
		resBytes, _ := json.Marshal(EnginesList{Engines: engines})
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
