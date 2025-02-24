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
	dataFlag        = "data"
	doneFlag        = []byte("[DONE]")
	errorFlag       = []byte(`"error"`)
	errorPrefixFlag = []byte(`{"error`)
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

func (stream *streamReader[T]) processLines() ([]byte, error) {
	var (
		emptyMessagesCount uint
		hasErrorPrefix     bool
	)

	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')
		if readErr != nil || (hasErrorPrefix && readErr == io.EOF) {
			respErr := stream.unmarshalError()
			if respErr != nil {
				return nil, fmt.Errorf("error, %w", respErr.Error)
			}
			return nil, readErr
		}

		// Split a string like "event: bar" into name="event" and value=" bar".
		name, value, _ := bytes.Cut(rawLine, []byte(":"))
		value = bytes.TrimSpace(value)

		// Consume an optional space after the colon if it exists.
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}

		switch string(name) {
		case dataFlag:
			if bytes.Equal(value, doneFlag) {
				stream.isFinished = true
				return nil, io.EOF
			}
			if bytes.HasPrefix(value, errorPrefixFlag) {
				if writeErr := stream.writeErrAccumulator(value); writeErr != nil {
					return nil, writeErr
				}
				respErr := stream.unmarshalError()
				if respErr != nil {
					return nil, fmt.Errorf("error, %w", respErr.Error)
				}
				continue
			}

			return value, nil
		default:
			if writeErr := stream.writeErrAccumulator(rawLine); writeErr != nil {
				return nil, writeErr
			}
			if bytes.Equal(name, errorFlag) {
				hasErrorPrefix = true
				continue
			}

			emptyMessagesCount++
			if emptyMessagesCount > stream.emptyMessagesLimit {
				return nil, ErrTooManyEmptyStreamMessages
			}

			continue
		}
	}
}

func (stream *streamReader[T]) writeErrAccumulator(p []byte) error {
	writeErr := stream.errAccumulator.Write(p)
	if writeErr != nil {
		return writeErr
	}

	return nil
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
