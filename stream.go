package openai

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrTooManyEmptyStreamMessages = errors.New("stream has sent too many empty messages")
)

type CompletionStream struct {
	*streamReader[CompletionResponse]
}

// CreateCompletionStream — API call to create a completion w/ streaming
// support. It sets whether to stream back partial progress. If set, tokens will be
// sent as data-only server-sent events as they become available, with the
// stream terminated by a data: [DONE] message.
func (c *Client) CreateCompletionStream(
	ctx context.Context,
	request CompletionRequest,
) (stream *CompletionStream, err error) {
	urlSuffix := "/completions"
	if !checkEndpointSupportsModel(urlSuffix, request.Model) {
		err = ErrCompletionUnsupportedModel
		return
	}

	if !checkPromptType(request.Prompt) {
		err = ErrCompletionRequestPromptTypeNotSupported
		return
	}

	request.Stream = true
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix, withModel(request.Model)),
		withBody(request),
	)
	if err != nil {
		return nil, err
	}

	resp, err := sendRequestStream[CompletionResponse](c, req)
	if err != nil {
		return
	}
	stream = &CompletionStream{
		streamReader: resp,
	}
	return
}
