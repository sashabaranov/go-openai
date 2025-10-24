package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	utils "github.com/sashabaranov/go-openai/internal"
)

var (
	headerData  = regexp.MustCompile(`^data:\s*`)
	errorPrefix = regexp.MustCompile(`^data:\s*{"error":`)
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
		// If we get a JSON parsing error, it might be because we got an error event
		// Check if we have accumulated error data
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) && len(stream.errAccumulator.Bytes()) > 0 {
			// We have error data, return a more informative error
			return response, fmt.Errorf("failed to parse response (error event received): %s",
				string(stream.errAccumulator.Bytes()))
		}
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
		emptyMessagesCount uint
		hasErrorPrefix     bool
	)

	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')
		if readErr != nil || hasErrorPrefix {
			respErr := stream.unmarshalError()
			if respErr != nil {
				return nil, respErr.Error
			}
			// If we detected an error event but couldn't parse it, and the stream ended,
			// return a more informative error. This handles cases where providers send
			// error events that don't match the expected format and immediately close.
			if hasErrorPrefix && readErr == io.EOF {
				// Check if we have error data that failed to parse
				errBytes := stream.errAccumulator.Bytes()
				if len(errBytes) > 0 {
					return nil, fmt.Errorf("failed to parse error event: %s", string(errBytes))
				}
				return nil, fmt.Errorf("stream ended after error event")
			}
			return nil, readErr
		}

		noSpaceLine := bytes.TrimSpace(rawLine)
		if errorPrefix.Match(noSpaceLine) {
			hasErrorPrefix = true
			// Extract just the JSON part after "data: " prefix
			// This handles both OpenAI format (data: {"error": ...}) and
			// Groq format (event: error\ndata: {"error": ...})
			jsonData := headerData.ReplaceAll(noSpaceLine, nil)
			writeErr := stream.errAccumulator.Write(jsonData)
			if writeErr != nil {
				return nil, writeErr
			}
			continue
		}

		// Skip non-data lines (e.g., "event: error" from Groq)
		// This allows us to handle SSE streams that use explicit event types
		if !headerData.Match(noSpaceLine) {
			emptyMessagesCount++
			if emptyMessagesCount > stream.emptyMessagesLimit {
				return nil, ErrTooManyEmptyStreamMessages
			}
			continue
		}

		noPrefixLine := headerData.ReplaceAll(noSpaceLine, nil)
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

	// Reset the error accumulator for future error events
	// A new accumulator is created to avoid potential interface issues
	stream.errAccumulator = utils.NewErrorAccumulator()

	return
}

func (stream *streamReader[T]) Close() error {
	return stream.response.Body.Close()
}
