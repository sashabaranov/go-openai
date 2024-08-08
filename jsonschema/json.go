// Package jsonschema provides very simple functionality for representing a JSON schema as a
// (nested) struct. This struct can be used with the chat completion "function call" feature.
// For more complicated schemas, it is recommended to use a dedicated JSON schema library
// and/or pass in the schema in []byte format.
package jsonschema

import (
	"encoding/json"
	"reflect"
	"strconv"
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
	Properties map[string]Definition `json:"properties,omitempty"`
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

func (d Definition) MarshalJSON() ([]byte, error) {
	if d.Properties == nil {
		d.Properties = make(map[string]Definition)
	}
	type Alias Definition
	return json.Marshal(struct {
		Alias
	}{
		Alias: (Alias)(d),
	})
}

type SchemaWrapper[T any] struct {
	data   T
	schema Definition
}

func (r SchemaWrapper[T]) Schema() Definition {
	return r.schema
}

func (r SchemaWrapper[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.schema)
}

func (r SchemaWrapper[T]) Unmarshal(content string) (*T, error) {
	var v T
	err := Unmarshal(r.schema, []byte(content), &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r SchemaWrapper[T]) String() string {
	bytes, _ := json.MarshalIndent(r.schema, "", "  ")
	return string(bytes)
}

func Warp[T any](v T) SchemaWrapper[T] {
	return SchemaWrapper[T]{
		data:   v,
		schema: reflectSchema(reflect.TypeOf(v)),
	}
}

func reflectSchema(t reflect.Type) Definition {
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
		items := reflectSchema(t.Elem())
		d.Items = &items
	case reflect.Struct:
		d.Type = Object
		d.AdditionalProperties = false
		properties := make(map[string]Definition)
		var requiredFields []string
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				jsonTag = field.Name
			}

			item := reflectSchema(field.Type)
			description := field.Tag.Get("description")
			if description != "" {
				item.Description = description
			}
			properties[jsonTag] = item

			required, _ := strconv.ParseBool(field.Tag.Get("required"))
			if required {
				requiredFields = append(requiredFields, jsonTag)
			}
		}
		d.Required = requiredFields
		d.Properties = properties
	default:
	}
	return d
}
