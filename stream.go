package gogpt

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

// CreateCompletionStream — API call to create a completion w/ streaming
// support. It sets whether to stream back partial progress. If set, tokens will be
// sent as data-only server-sent events as they become available, with the
// stream terminated by a data: [DONE] message.
func (c *Client) CreateCompletionStream(
	ctx context.Context,
	request CompletionRequest,
) ([]CompletionResponse, error) {
	request.Stream = true
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	urlSuffix := "/completions"
	req, err := http.NewRequest("POST", c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	reader := bufio.NewReader(res.Body)
	var line []byte
	var headerData = []byte("data: ")

	var responses []CompletionResponse
	for {
		line, err = reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Printf("error: %s", err)
		}

		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, headerData) {
			continue
		}

		line = bytes.TrimPrefix(line, headerData)
		if string(line) == "[DONE]" {
			responses = append(responses, CompletionResponse{ID: "[DONE]"})
			break
		}

		response := CompletionResponse{}
		err = json.Unmarshal(line, &response)
		if err != nil {
			log.Printf("invalid json stream data: %v", err)
		}
		responses = append(responses, response)
	}
	return responses, nil
}
