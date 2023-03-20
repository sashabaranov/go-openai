package openai

import (
	"encoding/json"
)

type marshaller interface {
	marshal(value any) ([]byte, error)
}

type jsonMarshaller struct{}

func (jm *jsonMarshaller) marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}
