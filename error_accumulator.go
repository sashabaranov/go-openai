package openai

import (
	"bytes"
	"fmt"
	"io"

	utils "github.com/sashabaranov/go-openai/internal"
)

type errorAccumulator interface {
	write(p []byte) error
	unmarshalError() *ErrorResponse
}

type errorBuffer interface {
	io.Writer
	Len() int
	Bytes() []byte
}

type defaultErrorAccumulator struct {
	buffer      errorBuffer
	unmarshaler utils.Unmarshaler
}

func newErrorAccumulator() errorAccumulator {
	return &defaultErrorAccumulator{
		buffer:      &bytes.Buffer{},
		unmarshaler: &utils.JSONUnmarshaler{},
	}
}

func (e *defaultErrorAccumulator) write(p []byte) error {
	_, err := e.buffer.Write(p)
	if err != nil {
		return fmt.Errorf("error accumulator write error, %w", err)
	}
	return nil
}

func (e *defaultErrorAccumulator) unmarshalError() (errResp *ErrorResponse) {
	if e.buffer.Len() == 0 {
		return
	}

	err := e.unmarshaler.Unmarshal(e.buffer.Bytes(), &errResp)
	if err != nil {
		errResp = nil
	}

	return
}
