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

// TestGetEngine Tests the retrieve engine endpoint of the API using the mocked server.
func TestGetEngine(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/engines/text-davinci-003", func(w http.ResponseWriter, r *http.Request) {
		resBytes, _ := json.Marshal(Engine{})
		fmt.Fprintln(w, string(resBytes))
	})
	// create the test server
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err := client.GetEngine(ctx, "text-davinci-003")
	checks.NoError(t, err, "GetEngine error")
}
