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
//
// On an HTTP-status error the returned stream is non-nil so its response headers
// (e.g. Retry-After, x-ratelimit-*) remain reachable via Header() and
// GetRateLimitHeaders(); Recv() on that stream returns io.EOF. On a transport-level
// failure (the request never reached the server) the returned stream is nil.
func (c *Client) CreateCompletionStream(
	ctx context.Context,
	request CompletionRequest,
) (stream *CompletionStream, err error) {
	urlSuffix := completionsSuffix
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
	if resp != nil {
		stream = &CompletionStream{
			streamReader: resp,
		}
	}
	return
}
