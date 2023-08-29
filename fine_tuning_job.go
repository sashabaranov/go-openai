package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type FineTuningJob struct {
	ID              string          `json:"id"`
	Object          string          `json:"object"`
	CreatedAt       int64           `json:"created_at"`
	FinishedAt      int64           `json:"finished_at"`
	Model           string          `json:"model"`
	FineTunedModel  string          `json:"fine_tuned_model,omitempty"`
	OrganizationID  string          `json:"organization_id"`
	Status          string          `json:"status"`
	Hyperparameters Hyperparameters `json:"hyperparameters"`
	TrainingFile    string          `json:"training_file"`
	ValidationFile  string          `json:"validation_file,omitempty"`
	ResultFiles     []string        `json:"result_files"`
	TrainedTokens   int             `json:"trained_tokens"`
}

type Hyperparameters struct {
	Epochs int `json:"n_epochs"`
}

type FineTuningJobRequest struct {
	TrainingFile    string           `json:"training_file"`
	ValidationFile  string           `json:"validation_file,omitempty"`
	Model           string           `json:"model,omitempty"`
	Hyperparameters *Hyperparameters `json:"hyperparameters,omitempty"`
	Suffix          string           `json:"suffix,omitempty"`
}

type FineTuningJobEventList struct {
	Object  string          `json:"object"`
	Data    []FineTuneEvent `json:"data"`
	HasMore bool            `json:"has_more"`
}

type FineTuningJobEvent struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	CreatedAt int    `json:"created_at"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	Type      string `json:"type"`
}

// CreateFineTuningJob create a fine tuning job.
func (c *Client) CreateFineTuningJob(
	ctx context.Context,
	request FineTuningJobRequest,
) (response FineTuningJob, err error) {
	urlSuffix := "/fine_tuning/jobs"
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CancelFineTuningJob cancel a fine tuning job.
func (c *Client) CancelFineTuningJob(ctx context.Context, fineTuningJobID string) (response FineTuningJob, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/fine_tuning/jobs/"+fineTuningJobID+"/cancel"))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveFineTuningJob retrieve a fine tuning job.
func (c *Client) RetrieveFineTuningJob(
	ctx context.Context,
	fineTuningJobID string,
) (response FineTuningJob, err error) {
	urlSuffix := fmt.Sprintf("/fine_tuning/jobs/%s", fineTuningJobID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

type listFineTuningJobEventsParameters struct {
	after *string
	limit *int
}

type ListFineTuningJobEventsParameter func(*listFineTuningJobEventsParameters)

func ListFineTuningJobEventsWithAfter(after string) ListFineTuningJobEventsParameter {
	return func(args *listFineTuningJobEventsParameters) {
		args.after = &after
	}
}

func ListFineTuningJobEventsWithLimit(limit int) ListFineTuningJobEventsParameter {
	return func(args *listFineTuningJobEventsParameters) {
		args.limit = &limit
	}
}

// ListFineTuningJobs list fine tuning jobs events.
func (c *Client) ListFineTuningJobEvents(
	ctx context.Context,
	fineTuningJobID string,
	setters ...ListFineTuningJobEventsParameter,
) (response FineTuningJobEventList, err error) {
	parameters := &listFineTuningJobEventsParameters{
		after: nil,
		limit: nil,
	}

	for _, setter := range setters {
		setter(parameters)
	}

	urlValues := url.Values{}
	if parameters.after != nil {
		urlValues.Add("after", *parameters.after)
	}
	if parameters.limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *parameters.limit))
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL("/fine_tuning/jobs/"+fineTuningJobID+"/events"+encodedValues),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
