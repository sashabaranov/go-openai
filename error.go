package openai

import (
	"encoding/json"
	"fmt"
)

// APIError provides error information returned by the OpenAI API.
type APIError struct {
	Code       json.RawMessage `json:"code,omitempty"`
	Message    string          `json:"message"`
	Param      *string         `json:"param,omitempty"`
	Type       string          `json:"type"`
	StatusCode int             `json:"-"`
}

// RequestError provides informations about generic request errors.
type RequestError struct {
	StatusCode int
	Err        error
}

type ErrorResponse struct {
	Error *APIError `json:"error,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) CodeAsStringPtr() (*string, error) {
	var s string
	if err := json.Unmarshal(e.Code, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (e *APIError) CodeAsInt() (int, error) {
	var i int
	if err := json.Unmarshal(e.Code, &i); err != nil {
		return 0, err
	}
	return i, nil
}

func (e *RequestError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("status code %d", e.StatusCode)
}

func (e *RequestError) Unwrap() error {
	return e.Err
}
