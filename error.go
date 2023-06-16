package openai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIError provides error information returned by the OpenAI API.
type APIError struct {
	Code           any     `json:"code,omitempty"`
	Message        string  `json:"message"`
	Param          *string `json:"param,omitempty"`
	Type           string  `json:"type"`
	HTTPStatusCode int     `json:"-"`
}

// RequestError provides informations about generic request errors.
type RequestError struct {
	HTTPStatusCode int
	Err            error
}

type ErrorResponse struct {
	Error *APIError `json:"error,omitempty"`
}

func (e *APIError) Error() string {
	if e.HTTPStatusCode > 0 {
		return fmt.Sprintf("error, status code: %d, message: %s", e.HTTPStatusCode, e.Message)
	}

	return e.Message
}

func (e *APIError) UnmarshalJSON(data []byte) (err error) {
	var rawMap map[string]json.RawMessage
	err = json.Unmarshal(data, &rawMap)
	if err != nil {
		return
	}

	err = json.Unmarshal(rawMap["message"], &e.Message)
	if err != nil {
		// If the parameter field of a function call is invalid as a JSON schema
		// refs: https://github.com/sashabaranov/go-openai/issues/381
		var messages []string
		err = json.Unmarshal(rawMap["message"], &messages)
		if err != nil {
			return
		}
		e.Message = strings.Join(messages, ", ")
	}

	// optional fields for azure openai
	// refs: https://github.com/sashabaranov/go-openai/issues/343
	if _, ok := rawMap["type"]; ok {
		err = json.Unmarshal(rawMap["type"], &e.Type)
		if err != nil {
			return
		}
	}

	// optional fields
	if _, ok := rawMap["param"]; ok {
		err = json.Unmarshal(rawMap["param"], &e.Param)
		if err != nil {
			return
		}
	}

	if _, ok := rawMap["code"]; !ok {
		return nil
	}

	// if the api returned a number, we need to force an integer
	// since the json package defaults to float64
	var intCode int
	err = json.Unmarshal(rawMap["code"], &intCode)
	if err == nil {
		e.Code = intCode
		return nil
	}

	return json.Unmarshal(rawMap["code"], &e.Code)
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("error, status code: %d, message: %s", e.HTTPStatusCode, e.Err)
}

func (e *RequestError) Unwrap() error {
	return e.Err
}
