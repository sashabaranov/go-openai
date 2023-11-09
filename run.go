package openai

import (
	"context"
	"fmt"
	"net/http"
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
	Type              RequiredActionType    `json:"type"`
	SubmitToolOutputs *RunSubmitToolOutputs `json:"submit_tool_outputs,omitempty"`
}

type RequiredActionType string

const (
	RequiredActionTypeSubmitToolOutputs RequiredActionType = "submit_tool_outputs"
)

type RunSubmitToolOutputs struct {
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
	ThreadID     string  `json:"-"`
	AssistantID  string  `json:"assistant_id"`
	Model        *string `json:"model,omitempty"`
	Instructions *string `json:"instructions,omitempty"`
	Tools        []Tool  `json:"tools,omitempty"`
	Metadata     map[string]any
}

// CreateRun creates a new run.
func (c *Client) CreateRun(ctx context.Context, request RunRequest) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/run", request.ThreadID)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
