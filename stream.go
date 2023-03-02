package gogpt

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

waitForData:
	line, err := stream.reader.ReadBytes('\n')
	if err != nil {
		return
	}

	var headerData = []byte("data: ")
	line = bytes.TrimSpace(line)
	if !bytes.HasPrefix(line, headerData) {
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
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return
	}

	urlSuffix := "/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))
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
