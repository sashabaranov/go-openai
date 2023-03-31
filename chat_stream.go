package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ChatCompletionStreamChoiceDelta struct {
	Content string `json:"content"`
}

type ChatCompletionStreamChoice struct {
	Index        int                             `json:"index"`
	Delta        ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason string                          `json:"finish_reason"`
}

type ChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
	Usage   Usage                        `json:"usage"`
}

// ChatCompletionStream
// Note: Perhaps it is more elegant to abstract Stream using generics.
type ChatCompletionStream struct {
	*streamReader[ChatCompletionStreamResponse]
}

// CreateChatCompletionStream — API call to create a chat completion w/ streaming
// support. It sets whether to stream back partial progress. If set, tokens will be
// sent as data-only server-sent events as they become available, with the
// stream terminated by a data: [DONE] message.
func (c *Client) CreateChatCompletionStream(
	ctx context.Context,
	request ChatCompletionRequest,
) (stream *ChatCompletionStream, err error) {
	urlSuffix := "/chat/completions"
	if !checkEndpointSupportsModel(urlSuffix, request.Model) {
		err = ErrChatCompletionInvalidModel
		return
	}

	request.Stream = true
	req, err := c.newStreamRequest(ctx, "POST", urlSuffix, request)
	if err != nil {
		return
	}

	res, err := c.config.HTTPClient.Do(req) //nolint:bodyclose // body is closed in stream.Close()
	if err != nil {
		return
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil || errRes.Error == nil {
			reqErr := RequestError{
				StatusCode: res.StatusCode,
				Err:        err,
			}
			err = fmt.Errorf("error, %w", &reqErr)
			return
		}
		errRes.Error.StatusCode = res.StatusCode
		err = fmt.Errorf("error, status code: %d, message: %w", res.StatusCode, errRes.Error)
		return
	}

	stream = &ChatCompletionStream{
		streamReader: &streamReader[ChatCompletionStreamResponse]{
			emptyMessagesLimit: c.config.EmptyMessagesLimit,
			reader:             bufio.NewReader(res.Body),
			response:           res,
			errAccumulator:     newErrorAccumulator(),
			unmarshaler:        &jsonUnmarshaler{},
		},
	}
	return
}
