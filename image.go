package openai

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"strconv"
)

// Image sizes defined by the OpenAI API.
const (
	CreateImageSize256x256   = "256x256"
	CreateImageSize512x512   = "512x512"
	CreateImageSize1024x1024 = "1024x1024"
	// dall-e-3 supported only.
	CreateImageSize1792x1024 = "1792x1024"
	CreateImageSize1024x1792 = "1024x1792"
)

const (
	CreateImageResponseFormatURL     = "url"
	CreateImageResponseFormatB64JSON = "b64_json"
)

const (
	CreateImageModelDallE2 = "dall-e-2"
	CreateImageModelDallE3 = "dall-e-3"
)

const (
	CreateImageQualityHD       = "hd"
	CreateImageQualityStandard = "standard"
)

const (
	CreateImageStyleVivid   = "vivid"
	CreateImageStyleNatural = "natural"
)

// ImageRequest represents the request structure for the image API.
type ImageRequest struct {
	Prompt         string `json:"prompt,omitempty"`
	Model          string `json:"model,omitempty"`
	N              int    `json:"n,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Size           string `json:"size,omitempty"`
	Style          string `json:"style,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

// ImageResponse represents a response structure for image API.
type ImageResponse struct {
	Created int64                    `json:"created,omitempty"`
	Data    []ImageResponseDataInner `json:"data,omitempty"`

	httpHeader
}

// ImageResponseDataInner represents a response data structure for image API.
type ImageResponseDataInner struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// CreateImage - API call to create an image. This is the main endpoint of the DALL-E API.
func (c *Client) CreateImage(ctx context.Context, request ImageRequest) (response ImageResponse, err error) {
	urlSuffix := "/images/generations"
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix, withModel(request.Model)),
		withBody(request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ImageEditRequest represents the request structure for the image API.
type ImageEditRequest struct {
	Image          *os.File `json:"image,omitempty"`
	Mask           *os.File `json:"mask,omitempty"`
	Prompt         string   `json:"prompt,omitempty"`
	Model          string   `json:"model,omitempty"`
	N              int      `json:"n,omitempty"`
	Size           string   `json:"size,omitempty"`
	ResponseFormat string   `json:"response_format,omitempty"`
}

// CreateEditImage - API call to create an image. This is the main endpoint of the DALL-E API.
func (c *Client) CreateEditImage(ctx context.Context, request ImageEditRequest) (response ImageResponse, err error) {
	body := &bytes.Buffer{}
	builder := c.createFormBuilder(body)

	// image
	err = builder.CreateFormFile("image", request.Image)
	if err != nil {
		return
	}

	// mask, it is optional
	if request.Mask != nil {
		err = builder.CreateFormFile("mask", request.Mask)
		if err != nil {
			return
		}
	}

	err = builder.WriteField("prompt", request.Prompt)
	if err != nil {
		return
	}

	err = builder.WriteField("n", strconv.Itoa(request.N))
	if err != nil {
		return
	}

	err = builder.WriteField("size", request.Size)
	if err != nil {
		return
	}

	err = builder.WriteField("response_format", request.ResponseFormat)
	if err != nil {
		return
	}

	err = builder.Close()
	if err != nil {
		return
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL("/images/edits", withModel(request.Model)),
		withBody(body),
		withContentType(builder.FormDataContentType()),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ImageVariRequest represents the request structure for the image API.
type ImageVariRequest struct {
	Image          *os.File `json:"image,omitempty"`
	Model          string   `json:"model,omitempty"`
	N              int      `json:"n,omitempty"`
	Size           string   `json:"size,omitempty"`
	ResponseFormat string   `json:"response_format,omitempty"`
}

// CreateVariImage - API call to create an image variation. This is the main endpoint of the DALL-E API.
// Use abbreviations(vari for variation) because ci-lint has a single-line length limit ...
func (c *Client) CreateVariImage(ctx context.Context, request ImageVariRequest) (response ImageResponse, err error) {
	body := &bytes.Buffer{}
	builder := c.createFormBuilder(body)

	// image
	err = builder.CreateFormFile("image", request.Image)
	if err != nil {
		return
	}

	err = builder.WriteField("n", strconv.Itoa(request.N))
	if err != nil {
		return
	}

	err = builder.WriteField("size", request.Size)
	if err != nil {
		return
	}

	err = builder.WriteField("response_format", request.ResponseFormat)
	if err != nil {
		return
	}

	err = builder.Close()
	if err != nil {
		return
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL("/images/variations", withModel(request.Model)),
		withBody(body),
		withContentType(builder.FormDataContentType()),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
