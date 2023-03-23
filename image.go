package openai

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

// Image sizes defined by the OpenAI API.
const (
	CreateImageSize256x256   = "256x256"
	CreateImageSize512x512   = "512x512"
	CreateImageSize1024x1024 = "1024x1024"
)

const (
	CreateImageResponseFormatURL     = "url"
	CreateImageResponseFormatB64JSON = "b64_json"
)

// ImageRequest represents the request structure for the image API.
type ImageRequest struct {
	Prompt         string `json:"prompt,omitempty"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

// ImageResponse represents a response structure for image API.
type ImageResponse struct {
	Created int64                    `json:"created,omitempty"`
	Data    []ImageResponseDataInner `json:"data,omitempty"`
}

// ImageResponseDataInner represents a response data structure for image API.
type ImageResponseDataInner struct {
	URL     string `json:"url,omitempty"`
	B64JSON string `json:"b64_json,omitempty"`
}

// CreateImage - API call to create an image. This is the main endpoint of the DALL-E API.
func (c *Client) CreateImage(ctx context.Context, request ImageRequest) (response ImageResponse, err error) {
	urlSuffix := "/images/generations"
	req, err := c.requestBuilder.build(ctx, http.MethodPost, c.fullURL(urlSuffix), request)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ImageEditRequest represents the request structure for the image API.
type ImageEditRequest struct {
	Image  *os.File `json:"image,omitempty"`
	Mask   *os.File `json:"mask,omitempty"`
	Prompt string   `json:"prompt,omitempty"`
	N      int      `json:"n,omitempty"`
	Size   string   `json:"size,omitempty"`
}

// CreateEditImage - API call to create an image. This is the main endpoint of the DALL-E API.
func (c *Client) CreateEditImage(ctx context.Context, request ImageEditRequest) (response ImageResponse, err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// image
	image, err := writer.CreateFormFile("image", request.Image.Name())
	if err != nil {
		return
	}
	_, err = io.Copy(image, request.Image)
	if err != nil {
		return
	}

	// mask, it is optional
	if request.Mask != nil {
		mask, err2 := writer.CreateFormFile("mask", request.Mask.Name())
		if err2 != nil {
			return
		}
		_, err = io.Copy(mask, request.Mask)
		if err != nil {
			return
		}
	}

	err = writer.WriteField("prompt", request.Prompt)
	if err != nil {
		return
	}
	err = writer.WriteField("n", strconv.Itoa(request.N))
	if err != nil {
		return
	}
	err = writer.WriteField("size", request.Size)
	if err != nil {
		return
	}
	writer.Close()
	urlSuffix := "/images/edits"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(urlSuffix), body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	err = c.sendRequest(req, &response)
	return
}

// ImageVariRequest represents the request structure for the image API.
type ImageVariRequest struct {
	Image *os.File `json:"image,omitempty"`
	N     int      `json:"n,omitempty"`
	Size  string   `json:"size,omitempty"`
}

// CreateVariImage - API call to create an image variation. This is the main endpoint of the DALL-E API.
// Use abbreviations(vari for variation) because ci-lint has a single-line length limit ...
func (c *Client) CreateVariImage(ctx context.Context, request ImageVariRequest) (response ImageResponse, err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// image
	image, err := writer.CreateFormFile("image", request.Image.Name())
	if err != nil {
		return
	}
	_, err = io.Copy(image, request.Image)
	if err != nil {
		return
	}

	err = writer.WriteField("n", strconv.Itoa(request.N))
	if err != nil {
		return
	}
	err = writer.WriteField("size", request.Size)
	if err != nil {
		return
	}
	writer.Close()
	//https://platform.openai.com/docs/api-reference/images/create-variation
	urlSuffix := "/images/variations"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(urlSuffix), body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	err = c.sendRequest(req, &response)
	return
}
