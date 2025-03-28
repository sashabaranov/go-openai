package openai

import (
	"encoding/json"
)

type Marshaller interface {
	Marshal(value any) ([]byte, error)
	Unmarshal(data []byte, value any) error
}

type JSONMarshaller struct{}

func (jm *JSONMarshaller) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (jm *JSONMarshaller) Unmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}
