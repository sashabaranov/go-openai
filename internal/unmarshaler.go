package openai

import (
	"encoding/json"
)

type Unmarshaler interface {
	Unmarshal(data []byte, v any) error
}

type JSONUnmarshaler struct{}

func (jm *JSONUnmarshaler) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
