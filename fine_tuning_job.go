package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// FineTuningJobList is a list of fine-tune jobs.
type FineTuningJobList struct {
	FineTuningJobs []FineTuningJob `json:"data"`
	HasMore        bool            `json:"has_more"`

	httpHeader
}

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

	httpHeader
}

type Hyperparameters struct {
	Epochs any `json:"n_epochs,omitempty"`
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

	httpHeader
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

const fineTuningJobSuffix = "/fine_tuning/jobs"

// CreateFineTuningJob create a fine tuning job.
func (c *Client) CreateFineTuningJob(
	ctx context.Context,
	request FineTuningJobRequest,
) (response FineTuningJob, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(fineTuningJobSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

type listFineTuningJobsParameters struct {
	after *string
	limit *int
}

type ListFineTuningJobsParameters func(*listFineTuningJobsParameters)

func ListFineTuningJobsWithAfter(after string) ListFineTuningJobsParameters {
	return func(args *listFineTuningJobsParameters) {
		args.after = &after
	}
}

func ListFineTuningJobsWithLimit(limit int) ListFineTuningJobsParameters {
	return func(args *listFineTuningJobsParameters) {
		args.limit = &limit
	}
}

// ListFineTuningJobs Lists the fine-tuning jobs.
func (c *Client) ListFineTuningJobs(
	ctx context.Context,
	setters ...ListFineTuningJobsParameters,
) (response FineTuningJobList, err error) {
	parameters := &listFineTuningJobsParameters{
		after: nil,
		limit: nil,
	}

	for _, setter := range setters {
		setter(parameters)
	}
	urlValues := url.Values{}
	if parameters.limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *parameters.limit))
	}
	if parameters.after != nil {
		urlValues.Add("after", *parameters.after)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("%s%s", fineTuningJobSuffix, encodedValues)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
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
