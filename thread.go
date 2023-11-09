package openai

import (
	"context"
	"net/http"
)

const (
	threadsSuffix = "/threads"
)

type Thread struct {
	ID        string         `json:"id"`
	Object    string         `json:"object"`
	CreatedAt int64          `json:"created_at"`
	Metadata  map[string]any `json:"metadata"`

	httpHeader
}

type ThreadRequest struct {
	Messages []ThreadMessage `json:"messages,omitempty"`
	Metadata map[string]any  `json:"metadata,omitempty"`
}

type ModifyThreadRequest struct {
	Metadata map[string]any `json:"metadata"`
}

type ThreadMessageRole string

const (
	ThreadMessageRoleUser ThreadMessageRole = "user"
)

type ThreadMessage struct {
	Role     ThreadMessageRole `json:"role"`
	Content  string            `json:"content"`
	FileIDs  []string          `json:"file_ids,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

type ThreadDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

// CreateThread creates a new thread.
func (c *Client) CreateThread(ctx context.Context, request ThreadRequest) (response Thread, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(threadsSuffix), withBody(request),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveThread retrieves a thread.
func (c *Client) RetrieveThread(ctx context.Context, threadID string) (response Thread, err error) {
	urlSuffix := threadsSuffix + "/" + threadID
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ModifyThread modifies a thread.
func (c *Client) ModifyThread(
	ctx context.Context,
	threadID string,
	request ModifyThreadRequest,
) (response Thread, err error) {
	urlSuffix := threadsSuffix + "/" + threadID
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// DeleteThread deletes a thread.
func (c *Client) DeleteThread(
	ctx context.Context,
	threadID string,
) (response ThreadDeleteResponse, err error) {
	urlSuffix := threadsSuffix + "/" + threadID
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
