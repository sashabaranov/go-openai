package openai

import (
	"encoding/json"
)

type unmarshaler interface {
	unmarshal(data []byte, v any) error
}

type jsonUnmarshaler struct{}

func (jm *jsonUnmarshaler) unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
