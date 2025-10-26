package openai

import (
	"context"
	"net/http"
)

const (
	responsesSuffix = "/responses"
)

type CreateResponseRequest struct {
	Model              string `json:"model"`
	Input              any    `json:"input"`
	Tools              []Tool `json:"tools,omitempty"`
	PreviousResponseID string `json:"previous_response_id,omitempty"`
}

type CreateResponseResponse struct {
	ID      string `json:"id"`
	Created int64  `json:"created_at"`
	Error   any    `json:"error,omitempty"`
	Output  []any  `json:"output"`
	Model   string `json:"model"`

	httpHeader
}

func (c *Client) CreateResponse(
	ctx context.Context,
	request CreateResponseRequest,
) (response CreateResponseResponse, err error) {
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(responsesSuffix),
		withBody(request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
