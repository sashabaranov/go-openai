package openai

import (
	"bytes"
	"fmt"
)

type errorAccumulator struct {
	buffer      bytes.Buffer
	unmarshaler unmarshaler
}

func newErrorAccumulator() *errorAccumulator {
	return &errorAccumulator{
		unmarshaler: &jsonUnmarshaler{},
	}
}

func (e *errorAccumulator) write(p []byte) (int, error) {
	n, err := e.buffer.Write(p)
	if err != nil {
		return n, fmt.Errorf("error accumulator write error, %w", err)
	}
	return n, nil
}

func (e *errorAccumulator) unmarshalError() (*ErrorResponse, error) {
	if e.buffer.Len() > 0 {
		var errRes ErrorResponse
		err := e.unmarshaler.unmarshal(e.buffer.Bytes(), &errRes)
		if err != nil {
			return nil, err
		}
		return &errRes, nil
	}
	return nil, nil
}
