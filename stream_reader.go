package openai

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type streamable interface {
	ChatCompletionStreamResponse | CompletionResponse
}

type streamReader[T streamable] struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator errorAccumulator
	unmarshaler    unmarshaler
}

func (stream *streamReader[T]) Recv() (response T, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}

	var emptyMessagesCount uint

waitForData:
	line, err := stream.reader.ReadBytes('\n')
	if err != nil {
		respErr := stream.errAccumulator.unmarshalError()
		if respErr != nil {
			err = fmt.Errorf("error, %w", respErr.Error)
		}
		return
	}

	var headerData = []byte("data: ")
	line = bytes.TrimSpace(line)
	if !bytes.HasPrefix(line, headerData) {
		if writeErr := stream.errAccumulator.write(line); writeErr != nil {
			err = writeErr
			return
		}
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

	err = stream.unmarshaler.unmarshal(line, &response)
	return
}

func (stream *streamReader[T]) Close() {
	stream.response.Body.Close()
}
