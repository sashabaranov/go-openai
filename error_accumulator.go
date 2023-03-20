package openai

import (
	"bytes"
	"fmt"
	"io"
)

type errorAccumulator interface {
	write(p []byte) error
	unmarshalError() (*ErrorResponse, error)
}

type errorBuffer interface {
	io.Writer
	Len() int
	Bytes() []byte
}

type errorAccumulate struct {
	buffer      errorBuffer
	unmarshaler unmarshaler
}

func newErrorAccumulator() errorAccumulator {
	return &errorAccumulate{
		buffer:      &bytes.Buffer{},
		unmarshaler: &jsonUnmarshaler{},
	}
}

func (e *errorAccumulate) write(p []byte) error {
	_, err := e.buffer.Write(p)
	if err != nil {
		return fmt.Errorf("error accumulator write error, %w", err)
	}
	return nil
}

func (e *errorAccumulate) unmarshalError() (*ErrorResponse, error) {
	var err error
	if e.buffer.Len() > 0 {
		var errRes ErrorResponse
		err = e.unmarshaler.unmarshal(e.buffer.Bytes(), &errRes)
		if err != nil {
			return nil, err
		}
		return &errRes, nil
	}
	return nil, err
}
