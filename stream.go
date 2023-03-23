package openai

import (
	"bufio"
	"context"
	"errors"
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

	request.Stream = true
	req, err := c.newStreamRequest(ctx, "POST", urlSuffix, request)
	if err != nil {
		return
	}

	resp, err := c.config.HTTPClient.Do(req) //nolint:bodyclose // body is closed in stream.Close()
	if err != nil {
		return
	}

	stream = &CompletionStream{
		streamReader: &streamReader[CompletionResponse]{
			emptyMessagesLimit: c.config.EmptyMessagesLimit,
			reader:             bufio.NewReader(resp.Body),
			response:           resp,
			errAccumulator:     newErrorAccumulator(),
			unmarshaler:        &jsonUnmarshaler{},
		},
	}
	return
}
