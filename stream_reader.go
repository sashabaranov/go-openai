package openai

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type streamReader struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator errorAccumulator
}

func (stream *streamReader) Recv() (line []byte, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}

	var emptyMessagesCount uint

waitForData:
	line, err = stream.reader.ReadBytes('\n')
	if err != nil {
		if errRes, _ := stream.errAccumulator.unmarshalError(); errRes != nil {
			err = fmt.Errorf("error, %w", errRes.Error)
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

	return
}

func (stream *streamReader) Close() {
	stream.response.Body.Close()
}
