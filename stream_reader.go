package openai

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	utils "github.com/sashabaranov/go-openai/internal"
)

var (
	headerData  = []byte("data: ")
	errorPrefix = []byte(`data: {"error":`)
	headerEvent = []byte("event: ")
)

type streamable interface {
	ChatCompletionStreamResponse | CompletionResponse | AssistantThreadRunStreamResponse
}

type streamReader[T streamable] struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator utils.ErrorAccumulator
	unmarshaler    utils.Unmarshaler

	event       string
	handlers    map[string]eventHandler[T]
	handlerCtx  context.Context
	handlerErr  error
	lastRawLine []byte

	httpHeader
}

type eventHandler[T streamable] func(T, []byte)

func (stream *streamReader[T]) On(topic string, handler eventHandler[T]) error {
	if len(topic) == 0 {
		return ErrStreamEventEmptyTopic
	}
	if stream.handlers == nil {
		stream.handlers = make(map[string]eventHandler[T])
	}
	stream.handlers[topic] = handler
	return nil
}

func (stream *streamReader[T]) Run() error {
	if stream.handlerCtx == nil {
		stream.event = "message" // default event for chat completion stream
		ctx, cancel := context.WithCancel(context.Background())
		stream.handlerCtx = ctx
		go func() {
			defer func() {
				if r := recover(); r != nil {
					stream.handlerErr = ErrStreamEventCallbackPanic
				}
				cancel()
			}()
			for {
				resp, err := stream.Recv()
				if err != nil {
					stream.handlerErr = err
					return
				}
				if callback, ok := stream.handlers[stream.event]; ok {
					callback(resp, stream.lastRawLine[len(headerData):])
				}
			}
		}()
	}
	<-stream.handlerCtx.Done()
	return stream.handlerErr
}

func (stream *streamReader[T]) Recv() (response T, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}

	response, err = stream.processLines()
	return
}

//nolint:gocognit
func (stream *streamReader[T]) processLines() (T, error) {
	var (
		emptyMessagesCount uint
		hasErrorPrefix     bool
	)

	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')
		if readErr != nil || hasErrorPrefix {
			respErr := stream.unmarshalError()
			if respErr != nil {
				return *new(T), fmt.Errorf("error, %w", respErr.Error)
			}
			return *new(T), readErr
		}

		noSpaceLine := bytes.TrimSpace(rawLine)
		stream.lastRawLine = noSpaceLine

		if bytes.HasPrefix(noSpaceLine, errorPrefix) {
			hasErrorPrefix = true
		}
		if !bytes.HasPrefix(noSpaceLine, headerData) || hasErrorPrefix {
			if hasErrorPrefix {
				noSpaceLine = bytes.TrimPrefix(noSpaceLine, headerData)
			}
			writeErr := stream.errAccumulator.Write(noSpaceLine)
			if writeErr != nil {
				return *new(T), writeErr
			}
			emptyMessagesCount++
			if emptyMessagesCount > stream.emptyMessagesLimit {
				return *new(T), ErrTooManyEmptyStreamMessages
			}
			// should optimize the code above for checking empty messages to better support stream events
			if bytes.HasPrefix(noSpaceLine, headerEvent) {
				stream.event = string(bytes.TrimPrefix(noSpaceLine, headerEvent))
			}
			continue
		}

		noPrefixLine := bytes.TrimPrefix(noSpaceLine, headerData)
		if string(noPrefixLine) == "[DONE]" {
			stream.isFinished = true
			return *new(T), io.EOF
		}

		var response T
		unmarshalErr := stream.unmarshaler.Unmarshal(noPrefixLine, &response)
		if unmarshalErr != nil {
			return *new(T), unmarshalErr
		}

		return response, nil
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
