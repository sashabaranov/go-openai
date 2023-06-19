package openai

import (
	"bytes"
	"fmt"
	"io"
)

type ErrorAccumulator interface {
	Write(p []byte) error
	Bytes() []byte
}

type errorBuffer interface {
	io.Writer
	Len() int
	Bytes() []byte
}

type DefaultErrorAccumulator struct {
	Buffer errorBuffer
}

func NewErrorAccumulator() ErrorAccumulator {
	return &DefaultErrorAccumulator{
		Buffer: &bytes.Buffer{},
	}
}

func (e *DefaultErrorAccumulator) Write(p []byte) error {
	_, err := e.Buffer.Write(p)
	if err != nil {
		return fmt.Errorf("error accumulator write error, %w", err)
	}
	return nil
}

func (e *DefaultErrorAccumulator) Bytes() (errBytes []byte) {
	if e.Buffer.Len() == 0 {
		return
	}
	errBytes = e.Buffer.Bytes()
	return
}
