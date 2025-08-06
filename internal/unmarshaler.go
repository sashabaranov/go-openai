package openai

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/bytedance/sonic"
)

type Unmarshaler interface {
	Unmarshal(data []byte, v any) error
}

type JSONUnmarshaler struct{}

func (jm *JSONUnmarshaler) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func UnmarshalExtraFields(typ reflect.Type, data []byte) (map[string]json.RawMessage, error) {
	m := make(map[string]json.RawMessage)
	if err := sonic.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type is not a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			delete(m, jsonTag)
		} else {
			if !field.IsExported() {
				continue
			}
			delete(m, field.Name)
		}
	}

	extra := make(map[string]json.RawMessage, len(m))
	for k, v := range m {
		extra[k] = v
	}

	return extra, nil
}
