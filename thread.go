package openai

import (
	"context"
	"net/http"
)

const (
	threadsSuffix = "/threads"
)

type Thread struct {
	ID       string         `json:"id"`
	Object   string         `json:"object"`
	Created  int64          `json:"created"`
	Metadata map[string]any `json:"metadata"`

	httpHeader
}

type ThreadRequest struct {
	Messages []ThreadMessage `json:"messages"`
}

type ModifyThreadRequest struct {
	Metadata map[string]any `json:"metadata"`
}

type ThreadMessage struct {
	Role     string         `json:"role"`
	Content  string         `json:"content"`
	FileIDs  []string       `json:"file_ids"`
	Metadata map[string]any `json:"metadata"`
}

// CreateThread creates a new thread.
func (c *Client) CreateThread(ctx context.Context, request ThreadRequest) (response Thread, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(threadsSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveThread retrieves a thread.
func (c *Client) RetrieveThread(ctx context.Context, threadID string) (response Thread, err error) {
	urlSuffix := threadsSuffix + "/" + threadID
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
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
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// DeleteThread deletes a thread.
func (c *Client) DeleteThread(ctx context.Context, threadID string) (err error) {
	urlSuffix := threadsSuffix + "/" + threadID
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, nil)
	return
}
