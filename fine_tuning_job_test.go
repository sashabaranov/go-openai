package openai_test

import (
	"context"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

const testFineTuninigJobID = "fine-tuning-job-id"

// TestFineTuningJob Tests the fine tuning job endpoint of the API using the mocked server.
func TestFineTuningJob(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler(
		"/v1/fine_tuning/jobs",
		func(w http.ResponseWriter, r *http.Request) {
			resBytes, _ := json.Marshal(FineTuningJob{
				Object:         "fine_tuning.job",
				ID:             testFineTuninigJobID,
				Model:          "davinci-002",
				CreatedAt:      1692661014,
				FinishedAt:     1692661190,
				FineTunedModel: "ft:davinci-002:my-org:custom_suffix:7q8mpxmy",
				OrganizationID: "org-123",
				ResultFiles:    []string{"file-abc123"},
				Status:         "succeeded",
				ValidationFile: "",
				TrainingFile:   "file-abc123",
				Hyperparameters: Hyperparameters{
					Epochs: "auto",
				},
				TrainedTokens: 5768,
			})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	server.RegisterHandler(
		"/fine_tuning/jobs/"+testFineTuninigJobID+"/cancel",
		func(w http.ResponseWriter, r *http.Request) {
			resBytes, _ := json.Marshal(FineTuningJob{})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	server.RegisterHandler(
		"/v1/fine_tuning/jobs/"+testFineTuninigJobID,
		func(w http.ResponseWriter, r *http.Request) {
			var resBytes []byte
			resBytes, _ = json.Marshal(FineTuningJob{})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	server.RegisterHandler(
		"/v1/fine_tuning/jobs/"+testFineTuninigJobID+"/events",
		func(w http.ResponseWriter, r *http.Request) {
			resBytes, _ := json.Marshal(FineTuningJobEventList{})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	ctx := context.Background()

	_, err := client.CreateFineTuningJob(ctx, FineTuningJobRequest{})
	checks.NoError(t, err, "CreateFineTuningJob error")

	_, err = client.CancelFineTuningJob(ctx, testFineTuninigJobID)
	checks.NoError(t, err, "CancelFineTuningJob error")

	_, err = client.RetrieveFineTuningJob(ctx, testFineTuninigJobID)
	checks.NoError(t, err, "RetrieveFineTuningJob error")

	_, err = client.ListFineTuningJobEvents(ctx, testFineTuninigJobID)
	checks.NoError(t, err, "ListFineTuningJobEvents error")

	_, err = client.ListFineTuningJobEvents(
		ctx,
		testFineTuninigJobID,
		ListFineTuningJobEventsWithAfter("last-event-id"),
	)
	checks.NoError(t, err, "ListFineTuningJobEvents error")

	_, err = client.ListFineTuningJobEvents(
		ctx,
		testFineTuninigJobID,
		ListFineTuningJobEventsWithLimit(10),
	)
	checks.NoError(t, err, "ListFineTuningJobEvents error")

	_, err = client.ListFineTuningJobEvents(
		ctx,
		testFineTuninigJobID,
		ListFineTuningJobEventsWithAfter("last-event-id"),
		ListFineTuningJobEventsWithLimit(10),
	)
	checks.NoError(t, err, "ListFineTuningJobEvents error")
}
