package openai

import (
	"context"
	"net/http"
)

const (
	responsesSuffix = "/responses"
)

// CreateResponseRequest represents a request structure for the Responses API.
type CreateResponseRequest struct {
	Model              string             `json:"model"`
	Input              any                `json:"input"`
	Tools              []Tool             `json:"tools,omitempty"`
	PreviousResponseID string             `json:"previous_response_id,omitempty"`
	Reasoning          *ResponseReasoning `json:"reasoning,omitempty"`
	ServiceTier        string             `json:"service_tier,omitempty"`
}

// ResponseReasoning represents reasoning configuration for the Responses API.
type ResponseReasoning struct {
	Effort          string `json:"effort,omitempty"`
	GenerateSummary string `json:"generate_summary,omitempty"`
	Summary         string `json:"summary,omitempty"`
}

// CreateResponseResponse represents a response structure for the Responses API.
type CreateResponseResponse struct {
	ID      string `json:"id"`
	Created int64  `json:"created_at"`
	Error   any    `json:"error,omitempty"`
	Output  []any  `json:"output"`
	Model   string `json:"model"`

	httpHeader
}

// CreateResponse creates a response using the Responses API.
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
