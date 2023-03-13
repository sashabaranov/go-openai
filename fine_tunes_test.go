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

const testFineTuneID = "fine-tune-id"

// TestFineTunes Tests the fine tunes endpoint of the API using the mocked server.
func TestFineTunes(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler(
		"/v1/fine-tunes",
		func(w http.ResponseWriter, r *http.Request) {
			var resBytes []byte
			if r.Method == http.MethodGet {
				resBytes, _ = json.Marshal(FineTuneList{})
			} else {
				resBytes, _ = json.Marshal(FineTune{})
			}
			fmt.Fprintln(w, string(resBytes))
		},
	)

	server.RegisterHandler(
		"/v1/fine-tunes/"+testFineTuneID+"/cancel",
		func(w http.ResponseWriter, r *http.Request) {
			resBytes, _ := json.Marshal(FineTune{})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	server.RegisterHandler(
		"/v1/fine-tunes/"+testFineTuneID,
		func(w http.ResponseWriter, r *http.Request) {
			var resBytes []byte
			if r.Method == http.MethodDelete {
				resBytes, _ = json.Marshal(FineTuneDeleteResponse{})
			} else {
				resBytes, _ = json.Marshal(FineTune{})
			}
			fmt.Fprintln(w, string(resBytes))
		},
	)

	server.RegisterHandler(
		"/v1/fine-tunes/"+testFineTuneID+"/events",
		func(w http.ResponseWriter, r *http.Request) {
			resBytes, _ := json.Marshal(FineTuneEventList{})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err = client.ListFineTunes(ctx)
	if err != nil {
		t.Fatalf("ListFineTunes error: %v", err)
	}

	_, err = client.CreateFineTune(ctx, FineTuneRequest{})
	if err != nil {
		t.Fatalf("CreateFineTune error: %v", err)
	}

	_, err = client.CancelFineTune(ctx, testFineTuneID)
	if err != nil {
		t.Fatalf("CancelFineTune error: %v", err)
	}

	_, err = client.GetFineTune(ctx, testFineTuneID)
	if err != nil {
		t.Fatalf("GetFineTune error: %v", err)
	}

	_, err = client.DeleteFineTune(ctx, testFineTuneID)
	if err != nil {
		t.Fatalf("DeleteFineTune error: %v", err)
	}

	_, err = client.ListFineTuneEvents(ctx, testFineTuneID)
	if err != nil {
		t.Fatalf("ListFineTuneEvents error: %v", err)
	}
}
