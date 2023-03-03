package gogpt

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CompletionChatStream struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader   *bufio.Reader
	response *http.Response
}

func (stream *CompletionChatStream) Recv() (response CompletionTurboResponse, err error) {
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

	_ = json.Unmarshal(line, &response)
	if response.Choices[0].FinishReason == "stop" {
		stream.isFinished = true
		err = io.EOF
		return
	}
	return
}

func (stream *CompletionChatStream) Close() {
	stream.response.Body.Close()
}

func (c *Client) CreateCompletionChatStream(
	ctx context.Context,
	request CompletionTurboRequestBody,
) (stream *CompletionChatStream, err error) {
	request.Stream = true
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return
	}
	urlSuffix := "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))
	if err != nil {
		return
	}
	var HTTPClient *http.Client
	HTTPClient = &http.Client{}

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return
	}

	stream = &CompletionChatStream{
		emptyMessagesLimit: 300,

		reader:   bufio.NewReader(resp.Body),
		response: resp,
	}
	return
}
