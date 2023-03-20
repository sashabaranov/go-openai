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

func (e *errorAccumulator) write(p []byte) error {
	_, err := e.buffer.Write(p)
	if err != nil {
		return fmt.Errorf("error accumulator write error, %w", err)
	}
	return nil
}

func (e *errorAccumulator) unmarshalError() (errRes *ErrorResponse, err error) {
	if e.buffer.Len() > 0 {
		errRes = &ErrorResponse{}
		err = e.unmarshaler.unmarshal(e.buffer.Bytes(), errRes)
		return
	}
	return
}
