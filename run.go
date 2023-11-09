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
	FileIDS        []string           `json:"file_ids"`
	Metadata       map[string]any     `json:"metadata"`

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
	AssistantID  string  `json:"assistant_id"`
	Model        *string `json:"model,omitempty"`
	Instructions *string `json:"instructions,omitempty"`
	Tools        []Tool  `json:"tools,omitempty"`
	Metadata     map[string]any
}

type RunModifyRequest struct {
	Metadata map[string]any `json:"metadata,omitempty"`
}

// RunList is a list of assistants.
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
	// Thread *ThreadRequest `json:"thread,omitempty"` uncomment when thread is implemented
}

// CreateRun creates a new run.
func (c *Client) CreateRun(
	ctx context.Context,
	threadID string,
	request RunRequest,
) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs", threadID)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantV1())
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
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantV1())
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
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantV1())
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
	limit *int,
	order *string,
	after *string,
	before *string,
) (response RunList, err error) {
	urlValues := url.Values{}
	if limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if order != nil {
		urlValues.Add("order", *order)
	}
	if after != nil {
		urlValues.Add("after", *after)
	}
	if before != nil {
		urlValues.Add("before", *before)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("/threads/%s/runs%s", threadID, encodedValues)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantV1())
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
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantV1())
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
	runID string,
	request SubmitToolOutputsRequest) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/cancel", threadID, runID)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix),
		withBetaAssistantV1())
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
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
