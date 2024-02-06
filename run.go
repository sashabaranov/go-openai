package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type Run struct {
	ID             string             `json:"id"`
	Object         string             `json:"object"`
	CreatedAt      int64              `json:"created_at"`
	ThreadID       string             `json:"thread_id"`
	AssistantID    string             `json:"assistant_id"`
	Status         RunStatus          `json:"status"`
	RequiredAction *RunRequiredAction `json:"required_action,omitempty"`
	LastError      *RunLastError      `json:"last_error,omitempty"`
	ExpiresAt      int64              `json:"expires_at"`
	StartedAt      *int64             `json:"started_at,omitempty"`
	CancelledAt    *int64             `json:"cancelled_at,omitempty"`
	FailedAt       *int64             `json:"failed_at,omitempty"`
	CompletedAt    *int64             `json:"completed_at,omitempty"`
	Model          string             `json:"model"`
	Instructions   string             `json:"instructions,omitempty"`
	Tools          []Tool             `json:"tools"`
	FileIDS        []string           `json:"file_ids"` //nolint:revive // backwards-compatibility
	Metadata       map[string]any     `json:"metadata"`
	Usage          Usage              `json:"usage,omitempty"`

	httpHeader
}

type RunStatus string

const (
	RunStatusQueued         RunStatus = "queued"
	RunStatusInProgress     RunStatus = "in_progress"
	RunStatusRequiresAction RunStatus = "requires_action"
	RunStatusCancelling     RunStatus = "cancelling"
	RunStatusFailed         RunStatus = "failed"
	RunStatusCompleted      RunStatus = "completed"
	RunStatusExpired        RunStatus = "expired"
	RunStatusCancelled      RunStatus = "cancelled"
)

type RunRequiredAction struct {
	Type              RequiredActionType `json:"type"`
	SubmitToolOutputs *SubmitToolOutputs `json:"submit_tool_outputs,omitempty"`
}

type RequiredActionType string

const (
	RequiredActionTypeSubmitToolOutputs RequiredActionType = "submit_tool_outputs"
)

type SubmitToolOutputs struct {
	ToolCalls []ToolCall `json:"tool_calls"`
}

type RunLastError struct {
	Code    RunError `json:"code"`
	Message string   `json:"message"`
}

type RunError string

const (
	RunErrorServerError       RunError = "server_error"
	RunErrorRateLimitExceeded RunError = "rate_limit_exceeded"
)

type RunRequest struct {
	AssistantID            string         `json:"assistant_id"`
	Model                  string         `json:"model,omitempty"`
	Instructions           string         `json:"instructions,omitempty"`
	AdditionalInstructions string         `json:"additional_instructions,omitempty"`
	Tools                  []Tool         `json:"tools,omitempty"`
	Metadata               map[string]any `json:"metadata,omitempty"`
}

type RunModifyRequest struct {
	Metadata map[string]any `json:"metadata,omitempty"`
}

// RunList is a list of runs.
type RunList struct {
	Runs []Run `json:"data"`

	httpHeader
}

type SubmitToolOutputsRequest struct {
	ToolOutputs []ToolOutput `json:"tool_outputs"`
}

type ToolOutput struct {
	ToolCallID string `json:"tool_call_id"`
	Output     any    `json:"output"`
}

type CreateThreadAndRunRequest struct {
	RunRequest
	Thread ThreadRequest `json:"thread"`
}

type RunStep struct {
	ID          string         `json:"id"`
	Object      string         `json:"object"`
	CreatedAt   int64          `json:"created_at"`
	AssistantID string         `json:"assistant_id"`
	ThreadID    string         `json:"thread_id"`
	RunID       string         `json:"run_id"`
	Type        RunStepType    `json:"type"`
	Status      RunStepStatus  `json:"status"`
	StepDetails StepDetails    `json:"step_details"`
	LastError   *RunLastError  `json:"last_error,omitempty"`
	ExpiredAt   *int64         `json:"expired_at,omitempty"`
	CancelledAt *int64         `json:"cancelled_at,omitempty"`
	FailedAt    *int64         `json:"failed_at,omitempty"`
	CompletedAt *int64         `json:"completed_at,omitempty"`
	Metadata    map[string]any `json:"metadata"`

	httpHeader
}

type RunStepStatus string

const (
	RunStepStatusInProgress RunStepStatus = "in_progress"
	RunStepStatusCancelling RunStepStatus = "cancelled"
	RunStepStatusFailed     RunStepStatus = "failed"
	RunStepStatusCompleted  RunStepStatus = "completed"
	RunStepStatusExpired    RunStepStatus = "expired"
)

type RunStepType string

const (
	RunStepTypeMessageCreation RunStepType = "message_creation"
	RunStepTypeToolCalls       RunStepType = "tool_calls"
)

type StepDetails struct {
	Type            RunStepType                 `json:"type"`
	MessageCreation *StepDetailsMessageCreation `json:"message_creation,omitempty"`
	ToolCalls       []ToolCall                  `json:"tool_calls,omitempty"`
}

type StepDetailsMessageCreation struct {
	MessageID string `json:"message_id"`
}

// RunStepList is a list of steps.
type RunStepList struct {
	RunSteps []RunStep `json:"data"`

	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
	HasMore bool   `json:"has_more"`

	httpHeader
}

type Pagination struct {
	Limit  *int
	Order  *string
	After  *string
	Before *string
}

// CreateRun creates a new run.
func (c *Client) CreateRun(
	ctx context.Context,
	threadID string,
	request RunRequest,
) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs", threadID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveRun retrieves a run.
func (c *Client) RetrieveRun(
	ctx context.Context,
	threadID string,
	runID string,
) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s", threadID, runID)
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(urlSuffix),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ModifyRun modifies a run.
func (c *Client) ModifyRun(
	ctx context.Context,
	threadID string,
	runID string,
	request RunModifyRequest,
) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s", threadID, runID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ListRuns lists runs.
func (c *Client) ListRuns(
	ctx context.Context,
	threadID string,
	pagination Pagination,
) (response RunList, err error) {
	urlValues := url.Values{}
	if pagination.Limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *pagination.Limit))
	}
	if pagination.Order != nil {
		urlValues.Add("order", *pagination.Order)
	}
	if pagination.After != nil {
		urlValues.Add("after", *pagination.After)
	}
	if pagination.Before != nil {
		urlValues.Add("before", *pagination.Before)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("/threads/%s/runs%s", threadID, encodedValues)
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(urlSuffix),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// SubmitToolOutputs submits tool outputs.
func (c *Client) SubmitToolOutputs(
	ctx context.Context,
	threadID string,
	runID string,
	request SubmitToolOutputsRequest) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/submit_tool_outputs", threadID, runID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CancelRun cancels a run.
func (c *Client) CancelRun(
	ctx context.Context,
	threadID string,
	runID string) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/cancel", threadID, runID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CreateThreadAndRun submits tool outputs.
func (c *Client) CreateThreadAndRun(
	ctx context.Context,
	request CreateThreadAndRunRequest) (response Run, err error) {
	urlSuffix := "/threads/runs"
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveRunStep retrieves a run step.
func (c *Client) RetrieveRunStep(
	ctx context.Context,
	threadID string,
	runID string,
	stepID string,
) (response RunStep, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/steps/%s", threadID, runID, stepID)
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(urlSuffix),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ListRunSteps lists run steps.
func (c *Client) ListRunSteps(
	ctx context.Context,
	threadID string,
	runID string,
	pagination Pagination,
) (response RunStepList, err error) {
	urlValues := url.Values{}
	if pagination.Limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *pagination.Limit))
	}
	if pagination.Order != nil {
		urlValues.Add("order", *pagination.Order)
	}
	if pagination.After != nil {
		urlValues.Add("after", *pagination.After)
	}
	if pagination.Before != nil {
		urlValues.Add("before", *pagination.Before)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/steps%s", threadID, runID, encodedValues)
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(urlSuffix),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
