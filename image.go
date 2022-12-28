package gogpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ImageSize string
type ResponseFormat string

const (
	ImageSizeSmall  ImageSize = "256x256"
	ImageSizeMedium ImageSize = "512x512"
	ImageSizeBig    ImageSize = "1024x1024"

	ImageResponseFormatURLs   ResponseFormat = "url"
	ImageResponseFormatBase64 ResponseFormat = "b64_json"

	ImageMaxPromptLength = 1000

	URLImageGeneration = "/images/generations"
)

type ImageCreateRequest struct {
	Prompt         string         `json:"prompt"`
	N              int            `json:"n,omitempty"`
	Size           ImageSize      `json:"size,omitempty"`
	ResponseFormat ResponseFormat `json:"response_format,omitempty"`
	User           string         `json:"user,omitempty"`
}

type ImageCreateResponse struct {
	CreatedAt time.Time
	URLs      []string
	Images    []string
}

func (resp *ImageCreateResponse) UnmarshalJSON(data []byte) error {
	var res struct {
		CreatedAt int64 `json:"created"`
		Data      []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &res); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	resp.CreatedAt = time.UnixMilli(res.CreatedAt)

	var urls []string
	var images []string

	for _, d := range res.Data {
		if d.URL != "" {
			urls = append(urls, d.URL)
		}
		if d.B64JSON != "" {
			images = append(images, d.B64JSON)
		}
	}

	resp.URLs = urls
	resp.Images = images

	return nil
}

func (c *Client) CreateImageCreate(
	ctx context.Context,
	request ImageCreateRequest,
) (*ImageCreateResponse, error) {

	if len(request.Prompt) > ImageMaxPromptLength {
		return nil, fmt.Errorf("prompt too long, max length is %d", ImageMaxPromptLength)
	}

	var reqBytes []byte
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.fullURL(URLImageGeneration), bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req = req.WithContext(ctx)

	var res ImageCreateResponse
	if err := c.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return &res, nil
}
