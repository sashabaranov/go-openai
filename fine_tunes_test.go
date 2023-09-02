package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

const testFineTuneID = "fine-tune-id"

// TestFineTunes Tests the fine tunes endpoint of the API using the mocked server.
func TestFineTunes(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
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

	ctx := context.Background()

	_, err := client.ListFineTunes(ctx)
	checks.NoError(t, err, "ListFineTunes error")

	_, err = client.CreateFineTune(ctx, FineTuneRequest{})
	checks.NoError(t, err, "CreateFineTune error")

	_, err = client.CancelFineTune(ctx, testFineTuneID)
	checks.NoError(t, err, "CancelFineTune error")

	_, err = client.GetFineTune(ctx, testFineTuneID)
	checks.NoError(t, err, "GetFineTune error")

	_, err = client.DeleteFineTune(ctx, testFineTuneID)
	checks.NoError(t, err, "DeleteFineTune error")

	_, err = client.ListFineTuneEvents(ctx, testFineTuneID)
	checks.NoError(t, err, "ListFineTuneEvents error")
}
