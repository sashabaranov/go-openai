package gogpt

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type ImageRequest struct {
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

type ImageResponse struct {
	Data []ImageData `json:"data"`
}

type ImageData struct {
	B64JSON string `json:"b64_json,omitempty"`
	URL     string `json:"url,omitempty"`
}

// Images creates an image given a prompt. See
// https://beta.openai.com/docs/api-reference/images/create.
func (c *Client) Images(ctx context.Context, request ImageRequest) (response ImageResponse, err error) {
	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", c.fullURL("/images/generations"), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &response)
	return
}
