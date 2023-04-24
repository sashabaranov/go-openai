package openai

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"strconv"
)

type (
	ImageSize           string
	ImageResponseFormat string
)

// Image suffix API path
const (
	_createImageSuffix = "/images/generations"
	_editImageSuffix   = "/images/edits"
	_createVariImage   = "/images/variations" // https://platform.openai.com/docs/api-reference/images/create-variation
)

// Image sizes defined by the OpenAI API.
const (
	ImageSize256x256   ImageSize = "256x256"
	ImageSize512x512   ImageSize = "512x512"
	ImageSize1024x1024 ImageSize = "1024x1024"
)

const (
	ImageResponseFormatURL     ImageResponseFormat = "url"
	ImageResponseFormatB64JSON ImageResponseFormat = "b64_json"
)

// ImageRequest represents the request structure for the image API.
type ImageRequest struct {
	Prompt         string              `json:"prompt,omitempty"`          // Prompt A text description of the desired image(s). The maximum length is 1000 characters
	N              int                 `json:"n,omitempty"`               // N The number of images to generate. Must be between 1 and 10
	Size           ImageSize           `json:"size,omitempty"`            // Size The size of the generated images. Must be one of 256x256, 512x512, or 1024x1024
	ResponseFormat ImageResponseFormat `json:"response_format,omitempty"` // ResponseFormat The format in which the generated images are returned. Must be one of url or b64_json
	User           string              `json:"user,omitempty"`            // User A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse
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
	req, err := c.requestBuilder.build(ctx, http.MethodPost, c.fullURL(_createImageSuffix), request)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ImageEditRequest represents the request structure for the image API.
type ImageEditRequest struct {
	Image          *os.File            `json:"image,omitempty"`
	Mask           *os.File            `json:"mask,omitempty"`
	Prompt         string              `json:"prompt,omitempty"`          // Prompt A text description of the desired image(s). The maximum length is 1000 characters
	N              int                 `json:"n,omitempty"`               // N The number of images to generate. Must be between 1 and 10
	Size           ImageSize           `json:"size,omitempty"`            // Size The size of the generated images. Must be one of 256x256, 512x512, or 1024x1024
	ResponseFormat ImageResponseFormat `json:"response_format,omitempty"` // ResponseFormat The format in which the generated images are returned. Must be one of url or b64_json
}

// CreateEditImage - API call to create an image. This is the main endpoint of the DALL-E API.
func (c *Client) CreateEditImage(ctx context.Context, request ImageEditRequest) (response ImageResponse, err error) {
	body := &bytes.Buffer{}
	builder := c.createFormBuilder(body)

	// image
	err = builder.createFormFile("image", request.Image)
	if err != nil {
		return
	}

	// mask, it is optional
	if request.Mask != nil {
		err = builder.createFormFile("mask", request.Mask)
		if err != nil {
			return
		}
	}

	err = builder.writeField("prompt", request.Prompt)
	if err != nil {
		return
	}

	err = builder.writeField("n", strconv.Itoa(request.N))
	if err != nil {
		return
	}

	err = builder.writeField("size", request.Size.String())
	if err != nil {
		return
	}

	err = builder.writeField("response_format", request.ResponseFormat.String())
	if err != nil {
		return
	}

	err = builder.close()
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(_editImageSuffix), body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", builder.formDataContentType())
	err = c.sendRequest(req, &response)
	return
}

// ImageVariRequest represents the request structure for the image API.
type ImageVariRequest struct {
	Image          *os.File            `json:"image,omitempty"`
	N              int                 `json:"n,omitempty"`               // N The number of images to generate. Must be between 1 and 10
	Size           ImageSize           `json:"size,omitempty"`            // Size The size of the generated images. Must be one of 256x256, 512x512, or 1024x1024
	ResponseFormat ImageResponseFormat `json:"response_format,omitempty"` // ResponseFormat The format in which the generated images are returned. Must be one of url or b64_json
}

// CreateVariImage - API call to create an image variation. This is the main endpoint of the DALL-E API.
// Use abbreviations(vari for variation) because ci-lint has a single-line length limit ...
func (c *Client) CreateVariImage(ctx context.Context, request ImageVariRequest) (response ImageResponse, err error) {
	body := &bytes.Buffer{}
	builder := c.createFormBuilder(body)

	// image
	err = builder.createFormFile("image", request.Image)
	if err != nil {
		return
	}

	err = builder.writeField("n", strconv.Itoa(request.N))
	if err != nil {
		return
	}

	err = builder.writeField("size", request.Size.String())
	if err != nil {
		return
	}

	err = builder.writeField("response_format", request.ResponseFormat.String())
	if err != nil {
		return
	}

	err = builder.close()
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(_createVariImage), body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", builder.formDataContentType())
	err = c.sendRequest(req, &response)
	return
}

func (s ImageSize) String() string {
	switch s {
	case ImageSize256x256:
		return "256x256"
	case ImageSize512x512:
		return "512x512"
	case ImageSize1024x1024:
		return "1024x1024"
	default:
		return "1024x1024"
	}
}

func (i ImageResponseFormat) String() string {
	switch i {
	case ImageResponseFormatURL:
		return "url"
	case ImageResponseFormatB64JSON:
		return "b64_json"
	default:
		return "url"
	}
}
