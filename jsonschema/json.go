// Package jsonschema provides very simple functionality for representing a JSON schema as a
// (nested) struct. This struct can be used with the chat completion "function call" feature.
// For more complicated schemas, it is recommended to use a dedicated JSON schema library
// and/or pass in the schema in []byte format.
package jsonschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type DataType string

const (
	Object  DataType = "object"
	Number  DataType = "number"
	Integer DataType = "integer"
	String  DataType = "string"
	Array   DataType = "array"
	Null    DataType = "null"
	Boolean DataType = "boolean"
)

// OrderedProperties represents an ordered map of property names to their definitions
type OrderedProperties struct {
	Keys   []string
	Values map[string]Definition
}

// Definition is a struct for describing a JSON Schema.
// It is fairly limited, and you may have better luck using a third-party library.
type Definition struct {
	// Type specifies the data type of the schema.
	Type DataType `json:"type,omitempty"`
	// Description is the description of the schema.
	Description string `json:"description,omitempty"`
	// Enum is used to restrict a value to a fixed set of values. It must be an array with at least
	// one element, where each element is unique. You will probably only use this with strings.
	Enum []string `json:"enum,omitempty"`
	// Properties describes the properties of an object, if the schema type is Object.
	Properties OrderedProperties `json:"-"`
	// Required specifies which properties are required, if the schema type is Object.
	Required []string `json:"required,omitempty"`
	// Items specifies which data type an array contains, if the schema type is Array.
	Items *Definition `json:"items,omitempty"`
	// AdditionalProperties is used to control the handling of properties in an object
	// that are not explicitly defined in the properties section of the schema. example:
	// additionalProperties: true
	// additionalProperties: false
	// additionalProperties: jsonschema.Definition{Type: jsonschema.String}
	AdditionalProperties any `json:"additionalProperties,omitempty"`
}

func (d *Definition) MarshalJSON() ([]byte, error) {
	type Alias Definition
	aux := struct {
		*Alias
		Properties json.RawMessage `json:"properties,omitempty"`
	}{
		Alias: (*Alias)(d),
	}

	if len(d.Properties.Keys) > 0 {
		orderedProps := make(map[string]json.RawMessage)
		for _, key := range d.Properties.Keys {
			value, err := json.Marshal(d.Properties.Values[key])
			if err != nil {
				return nil, err
			}
			orderedProps[key] = value
		}
		var err error
		aux.Properties, err = json.Marshal(orderedProps)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(aux)
}

func (d *Definition) Unmarshal(content string, v any) error {
	return VerifySchemaAndUnmarshal(*d, []byte(content), v)
}

func GenerateSchemaForType(v any) (*Definition, error) {
	return reflectSchema(reflect.TypeOf(v))
}

func reflectSchema(t reflect.Type) (*Definition, error) {
	var d Definition
	switch t.Kind() {
	case reflect.String:
		d.Type = String
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		d.Type = Integer
	case reflect.Float32, reflect.Float64:
		d.Type = Number
	case reflect.Bool:
		d.Type = Boolean
	case reflect.Slice, reflect.Array:
		d.Type = Array
		items, err := reflectSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		d.Items = items
	case reflect.Struct:
		d.Type = Object
		d.AdditionalProperties = false
		object, err := reflectSchemaObject(t)
		if err != nil {
			return nil, err
		}
		d = *object
	case reflect.Ptr:
		definition, err := reflectSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		d = *definition
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.UnsafePointer:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind().String())
	default:
	}
	return &d, nil
}

func reflectSchemaObject(t reflect.Type) (*Definition, error) {
	var d = Definition{
		Type:                 Object,
		AdditionalProperties: false,
		Properties:           OrderedProperties{Keys: make([]string, 0), Values: make(map[string]Definition)},
	}
	var requiredFields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		jsonTag := field.Tag.Get("json")
		var required = true
		if jsonTag == "" {
			jsonTag = field.Name
		} else if strings.HasSuffix(jsonTag, ",omitempty") {
			jsonTag = strings.TrimSuffix(jsonTag, ",omitempty")
			required = false
		}

		item, err := reflectSchema(field.Type)
		if err != nil {
			return nil, err
		}
		description := field.Tag.Get("description")
		if description != "" {
			item.Description = description
		}
		d.Properties.Keys = append(d.Properties.Keys, jsonTag)
		d.Properties.Values[jsonTag] = *item

		if s := field.Tag.Get("required"); s != "" {
			required, _ = strconv.ParseBool(s)
		}
		if required {
			requiredFields = append(requiredFields, jsonTag)
		}
	}
	d.Required = requiredFields
	return &d, nil
}
