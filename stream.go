package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
)

var (
	ErrTooManyEmptyStreamMessages = errors.New("stream has sent too many empty messages")
)

type CompletionStream struct {
	*streamReader
}

func (stream *CompletionStream) Recv() (response CompletionResponse, err error) {
	line, err := stream.streamReader.Recv()
	if err != nil {
		return
	}

	err = json.Unmarshal(line, &response)
	return
}

func (stream *CompletionStream) Close() {
	stream.streamReader.Close()
}

// CreateCompletionStream — API call to create a completion w/ streaming
// support. It sets whether to stream back partial progress. If set, tokens will be
// sent as data-only server-sent events as they become available, with the
// stream terminated by a data: [DONE] message.
func (c *Client) CreateCompletionStream(
	ctx context.Context,
	request CompletionRequest,
) (stream *CompletionStream, err error) {
	request.Stream = true
	req, err := c.newStreamRequest(ctx, "POST", "/completions", request)
	if err != nil {
		return
	}

	resp, err := c.config.HTTPClient.Do(req) //nolint:bodyclose // body is closed in stream.Close()
	if err != nil {
		return
	}

	stream = &CompletionStream{
		streamReader: &streamReader{
			emptyMessagesLimit: c.config.EmptyMessagesLimit,
			reader:             bufio.NewReader(resp.Body),
			response:           resp,
			errAccumulator:     newErrorAccumulator(),
		},
	}
	return
}
