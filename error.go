package openai

import (
	"encoding/json"
	"fmt"
)

// APIError provides error information returned by the OpenAI API.
type APIError struct {
	C          json.RawMessage `json:"code,omitempty"`
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

// Code returns the error code as an int or string, depending on the API response.
func (e *APIError) Code() (any, error) {
	var i int
	if err := json.Unmarshal(e.C, &i); err == nil {
		return i, nil
	}
	var s string
	if err := json.Unmarshal(e.C, &s); err == nil {
		return &s, nil
	}
	return nil, fmt.Errorf("unknown code type: %s", e.C)
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
