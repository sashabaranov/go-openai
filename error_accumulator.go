package openai

import (
	"bytes"
	"fmt"
	"io"
)

type errorAccumulator interface {
	write(p []byte) error
	bytes() []byte
}

type errorBuffer interface {
	io.Writer
	Len() int
	Bytes() []byte
}

type defaultErrorAccumulator struct {
	buffer errorBuffer
}

func newErrorAccumulator() errorAccumulator {
	return &defaultErrorAccumulator{
		buffer: &bytes.Buffer{},
	}
}

func (e *defaultErrorAccumulator) write(p []byte) error {
	_, err := e.buffer.Write(p)
	if err != nil {
		return fmt.Errorf("error accumulator write error, %w", err)
	}
	return nil
}

func (e *defaultErrorAccumulator) bytes() (errBytes []byte) {
	if e.buffer.Len() == 0 {
		return
	}
	errBytes = e.buffer.Bytes()
	return
}
