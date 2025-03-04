package openai

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"

	utils "github.com/sashabaranov/go-openai/internal"
)

var (
	dataField   = []byte("data")
	errorPrefix = []byte(`{"error":`)
)

const (
	splitParts = 2
)

type streamable interface {
	ChatCompletionStreamResponse | CompletionResponse
}

type streamReader[T streamable] struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator utils.ErrorAccumulator
	unmarshaler    utils.Unmarshaler

	httpHeader
}

func (stream *streamReader[T]) Recv() (response T, err error) {
	rawLine, err := stream.RecvRaw()
	if err != nil {
		return
	}

	err = stream.unmarshaler.Unmarshal(rawLine, &response)
	if err != nil {
		return
	}
	return response, nil
}

func (stream *streamReader[T]) RecvRaw() ([]byte, error) {
	if stream.isFinished {
		return nil, io.EOF
	}

	return stream.processLines()
}

//nolint:gocognit
func (stream *streamReader[T]) processLines() ([]byte, error) {
	var (
		emptyMessagesCount  uint
		dataFieldNotFound   bool
		valueHasErrorPrefix bool
	)

	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')
		if readErr != nil || valueHasErrorPrefix {
			respErr := stream.unmarshalError()
			if respErr != nil {
				return nil, fmt.Errorf("error, %w", respErr.Error)
			}
			return nil, readErr
		}

		noSpaceLine := bytes.TrimSpace(rawLine)

		var value []byte

		split := bytes.SplitN(noSpaceLine, []byte(":"), splitParts)

		if len(split) != splitParts || !bytes.Equal(split[0], dataField) {
			dataFieldNotFound = true
		} else {
			value = split[1]

			if bytes.HasPrefix(value, []byte(" ")) {
				value = value[1:]
			}

			if bytes.HasPrefix(value, errorPrefix) {
				valueHasErrorPrefix = true
			}
		}

		if dataFieldNotFound || valueHasErrorPrefix {
			if valueHasErrorPrefix {
				noSpaceLine = value
			}
			writeErr := stream.errAccumulator.Write(noSpaceLine)
			if writeErr != nil {
				return nil, writeErr
			}
			emptyMessagesCount++
			if emptyMessagesCount > stream.emptyMessagesLimit {
				return nil, ErrTooManyEmptyStreamMessages
			}
			dataFieldNotFound = false

			continue
		}

		noPrefixLine := value
		if string(noPrefixLine) == "[DONE]" {
			stream.isFinished = true
			return nil, io.EOF
		}

		return noPrefixLine, nil
	}
}

func (stream *streamReader[T]) unmarshalError() (errResp *ErrorResponse) {
	errBytes := stream.errAccumulator.Bytes()
	if len(errBytes) == 0 {
		return
	}

	err := stream.unmarshaler.Unmarshal(errBytes, &errResp)
	if err != nil {
		errResp = nil
	}

	return
}

func (stream *streamReader[T]) Close() error {
	return stream.response.Body.Close()
}
