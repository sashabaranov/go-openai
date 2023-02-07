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

type CompletionStream struct {
	reader   *bufio.Reader
	response *http.Response
}

func (stream *CompletionStream) Recv() (response CompletionResponse, err error) {
waitForData:
	line, err := stream.reader.ReadBytes('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
	}

	var headerData = []byte("data: ")
	line = bytes.TrimSpace(line)
	if !bytes.HasPrefix(line, headerData) {
		goto waitForData
	}

	line = bytes.TrimPrefix(line, headerData)
	if string(line) == "[DONE]" {
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
	req, err := http.NewRequest("POST", c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	resp, err := c.HTTPClient.Do(req) //nolint:bodyclose // body is closed in stream.Close()
	if err != nil {
		return
	}

	stream = &CompletionStream{
		reader:   bufio.NewReader(resp.Body),
		response: resp,
	}
	return
}
