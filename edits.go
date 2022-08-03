package gogpt

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// EditsRequest represents a request structure for Edits API.
type EditsRequest struct {
	Model       *string `json:"model,omitempty"`
	Input       string  `json:"input,omitempty"`
	Instruction string  `json:"instruction,omitempty"`
	N           int     `json:"n,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
	TopP        float32 `json:"top_p,omitempty"`
}

// EditsChoice represents one of possible edits.
type EditsChoice struct {
	Text  string `json:"text"`
	Index int    `json:"index"`
}

// EditsResponse represents a response structure for Edits API.
type EditsResponse struct {
	Object  string        `json:"object"`
	Created uint64        `json:"created"`
	Usage   Usage         `json:"usage"`
	Choices []EditsChoice `json:"choices"`
}

// Perform an API call to the Edits endpoint.
func (c *Client) Edits(ctx context.Context, request EditsRequest) (response EditsResponse, err error) {
	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", c.fullURL("/edits"), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &response)
	return
}
