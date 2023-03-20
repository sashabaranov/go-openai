package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrTooManyEmptyStreamMessages = errors.New("stream has sent too many empty messages")
)

type CompletionStream struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader   *bufio.Reader
	response *http.Response
}

func (stream *CompletionStream) Recv() (response CompletionResponse, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}

	var emptyMessagesCount uint
	var errBytes bytes.Buffer

waitForData:
	line, err := stream.reader.ReadBytes('\n')
	if err != nil {
		if errBytes.Len() > 0 {
			var errRes ErrorResponse
			if jsonErr := json.Unmarshal(errBytes.Bytes(), &errRes); jsonErr == nil {
				err = fmt.Errorf("error, %w", errRes.Error)
			}
		}
		return
	}

	var headerData = []byte("data: ")
	line = bytes.TrimSpace(line)
	if !bytes.HasPrefix(line, headerData) {
		errBytes.Write(line)
		emptyMessagesCount++
		if emptyMessagesCount > stream.emptyMessagesLimit {
			err = ErrTooManyEmptyStreamMessages
			return
		}

		goto waitForData
	}

	line = bytes.TrimPrefix(line, headerData)
	if string(line) == "[DONE]" {
		stream.isFinished = true
		err = io.EOF
		return
	}

	err = json.Unmarshal(line, &response)
	return
}

func (stream *CompletionStream) Close() {
	stream.response.Body.Close()
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
		emptyMessagesLimit: c.config.EmptyMessagesLimit,

		reader:   bufio.NewReader(resp.Body),
		response: resp,
	}
	return
}
