package openai

import (
	"bytes"
	"context"
	"io"
	"net/http"
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

	// gpt-image-1 supported only.
	CreateImageSize1536x1024 = "1536x1024" // Landscape
	CreateImageSize1024x1536 = "1024x1536" // Portrait
)

const (
	// dall-e-2 and dall-e-3 only.
	CreateImageResponseFormatB64JSON = "b64_json"
	CreateImageResponseFormatURL     = "url"
)

const (
	CreateImageModelDallE2    = "dall-e-2"
	CreateImageModelDallE3    = "dall-e-3"
	CreateImageModelGptImage1 = "gpt-image-1"
)

const (
	CreateImageQualityHD       = "hd"
	CreateImageQualityStandard = "standard"

	// gpt-image-1 only.
	CreateImageQualityHigh   = "high"
	CreateImageQualityMedium = "medium"
	CreateImageQualityLow    = "low"
)

const (
	// dall-e-3 only.
	CreateImageStyleVivid   = "vivid"
	CreateImageStyleNatural = "natural"
)

const (
	// gpt-image-1 only.
	CreateImageBackgroundTransparent = "transparent"
	CreateImageBackgroundOpaque      = "opaque"
)

const (
	// gpt-image-1 only.
	CreateImageModerationLow = "low"
)

const (
	// gpt-image-1 only.
	CreateImageOutputFormatPNG  = "png"
	CreateImageOutputFormatJPEG = "jpeg"
	CreateImageOutputFormatWEBP = "webp"
)

// ImageRequest represents the request structure for the image API.
type ImageRequest struct {
	Prompt            string `json:"prompt,omitempty"`
	Model             string `json:"model,omitempty"`
	N                 int    `json:"n,omitempty"`
	Quality           string `json:"quality,omitempty"`
	Size              string `json:"size,omitempty"`
	Style             string `json:"style,omitempty"`
	ResponseFormat    string `json:"response_format,omitempty"`
	User              string `json:"user,omitempty"`
	Background        string `json:"background,omitempty"`
	Moderation        string `json:"moderation,omitempty"`
	OutputCompression int    `json:"output_compression,omitempty"`
	OutputFormat      string `json:"output_format,omitempty"`
}

// ImageResponse represents a response structure for image API.
type ImageResponse struct {
	Created int64                    `json:"created,omitempty"`
	Data    []ImageResponseDataInner `json:"data,omitempty"`
	Usage   ImageResponseUsage       `json:"usage,omitempty"`

	httpHeader
}

// ImageResponseInputTokensDetails represents the token breakdown for input tokens.
type ImageResponseInputTokensDetails struct {
	TextTokens  int `json:"text_tokens,omitempty"`
	ImageTokens int `json:"image_tokens,omitempty"`
}

// ImageResponseUsage represents the token usage information for image API.
type ImageResponseUsage struct {
	TotalTokens        int                             `json:"total_tokens,omitempty"`
	InputTokens        int                             `json:"input_tokens,omitempty"`
	OutputTokens       int                             `json:"output_tokens,omitempty"`
	InputTokensDetails ImageResponseInputTokensDetails `json:"input_tokens_details,omitempty"`
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

// WrapReader wraps an io.Reader with filename and Content-type.
func WrapReader(rdr io.Reader, filename string, contentType string) io.Reader {
	return file{rdr, filename, contentType}
}

type file struct {
	io.Reader
	name        string
	contentType string
}

func (f file) Name() string {
	if f.name != "" {
		return f.name
	} else if named, ok := f.Reader.(interface{ Name() string }); ok {
		return named.Name()
	}
	return ""
}

func (f file) ContentType() string {
	return f.contentType
}

// ImageEditRequest represents the request structure for the image API.
// Use WrapReader to wrap an io.Reader with filename and Content-type.
type ImageEditRequest struct {
	Image          io.Reader `json:"image,omitempty"`
	Mask           io.Reader `json:"mask,omitempty"`
	Prompt         string    `json:"prompt,omitempty"`
	Model          string    `json:"model,omitempty"`
	N              int       `json:"n,omitempty"`
	Size           string    `json:"size,omitempty"`
	ResponseFormat string    `json:"response_format,omitempty"`
	Quality        string    `json:"quality,omitempty"`
	User           string    `json:"user,omitempty"`
}

// CreateEditImage - API call to create an image. This is the main endpoint of the DALL-E API.
func (c *Client) CreateEditImage(ctx context.Context, request ImageEditRequest) (response ImageResponse, err error) {
	body := &bytes.Buffer{}
	builder := c.createFormBuilder(body)

	// image, filename verification can be postponed
	err = builder.CreateFormFileReader("image", request.Image, "")
	if err != nil {
		return
	}

	// mask, it is optional
	if request.Mask != nil {
		// filename verification can be postponed
		err = builder.CreateFormFileReader("mask", request.Mask, "")
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
// Use WrapReader to wrap an io.Reader with filename and Content-type.
type ImageVariRequest struct {
	Image          io.Reader `json:"image,omitempty"`
	Model          string    `json:"model,omitempty"`
	N              int       `json:"n,omitempty"`
	Size           string    `json:"size,omitempty"`
	ResponseFormat string    `json:"response_format,omitempty"`
	User           string    `json:"user,omitempty"`
}

// CreateVariImage - API call to create an image variation. This is the main endpoint of the DALL-E API.
// Use abbreviations(vari for variation) because ci-lint has a single-line length limit ...
func (c *Client) CreateVariImage(ctx context.Context, request ImageVariRequest) (response ImageResponse, err error) {
	body := &bytes.Buffer{}
	builder := c.createFormBuilder(body)

	// image, filename verification can be postponed
	err = builder.CreateFormFileReader("image", request.Image, "")
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
