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
	// Whether the schema is nullable or not.
	Nullable bool `json:"nullable,omitempty"`

	// Ref Reference to a definition in $defs or external schema.
	Ref string `json:"$ref,omitempty"`
	// Defs A map of reusable schema definitions.
	Defs map[string]Definition `json:"$defs,omitempty"`
}

func (d *Definition) MarshalJSON() ([]byte, error) {
	if d.Properties == nil {
		d.Properties = make(map[string]Definition)
	}
	type Alias Definition
	return json.Marshal(struct {
		Alias
	}{
		Alias: (Alias)(*d),
	})
}

func (d *Definition) Unmarshal(content string, v any) error {
	return VerifySchemaAndUnmarshal(*d, []byte(content), v)
}

func GenerateSchemaForType(v any) (*Definition, error) {
	var defs = make(map[string]Definition)
	def, err := reflectSchema(reflect.TypeOf(v), defs)
	if err != nil {
		return nil, err
	}
	// If the schema has a root $ref, resolve it by:
	// 1. Extracting the key from the $ref.
	// 2. Detaching the referenced definition from $defs.
	// 3. Checking for self-references in the detached definition.
	//    - If a self-reference is found, restore the original $defs structure.
	// 4. Flattening the referenced definition into the root schema.
	// 5. Clearing the $ref field in the root schema.
	if def.Ref != "" {
		origRef := def.Ref
		key := strings.TrimPrefix(origRef, "#/$defs/")
		if root, ok := defs[key]; ok {
			delete(defs, key)
			root.Defs = defs
			if containsRef(root, origRef) {
				root.Defs = nil
				defs[key] = root
			}
			*def = root
		}
		def.Ref = ""
	}
	def.Defs = defs
	return def, nil
}

func reflectSchema(t reflect.Type, defs map[string]Definition) (*Definition, error) {
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
		items, err := reflectSchema(t.Elem(), defs)
		if err != nil {
			return nil, err
		}
		d.Items = items
	case reflect.Struct:
		if t.Name() != "" {
			if _, ok := defs[t.Name()]; !ok {
				defs[t.Name()] = Definition{}
				object, err := reflectSchemaObject(t, defs)
				if err != nil {
					return nil, err
				}
				defs[t.Name()] = *object
			}
			return &Definition{Ref: "#/$defs/" + t.Name()}, nil
		}
		d.Type = Object
		d.AdditionalProperties = false
		object, err := reflectSchemaObject(t, defs)
		if err != nil {
			return nil, err
		}
		d = *object
	case reflect.Ptr:
		definition, err := reflectSchema(t.Elem(), defs)
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

func reflectSchemaObject(t reflect.Type, defs map[string]Definition) (*Definition, error) {
	var d = Definition{
		Type:                 Object,
		AdditionalProperties: false,
	}
	properties := make(map[string]Definition)
	var requiredFields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		jsonTag := field.Tag.Get("json")
		var required = true
		switch {
		case jsonTag == "-":
			continue
		case jsonTag == "":
			jsonTag = field.Name
		case strings.HasSuffix(jsonTag, ",omitempty"):
			jsonTag = strings.TrimSuffix(jsonTag, ",omitempty")
			required = false
		}

		item, err := reflectSchema(field.Type, defs)
		if err != nil {
			return nil, err
		}
		description := field.Tag.Get("description")
		if description != "" {
			item.Description = description
		}
		enum := field.Tag.Get("enum")
		if enum != "" {
			item.Enum = strings.Split(enum, ",")
		}

		if n := field.Tag.Get("nullable"); n != "" {
			nullable, _ := strconv.ParseBool(n)
			item.Nullable = nullable
		}

		properties[jsonTag] = *item

		if s := field.Tag.Get("required"); s != "" {
			required, _ = strconv.ParseBool(s)
		}
		if required {
			requiredFields = append(requiredFields, jsonTag)
		}
	}
	d.Required = requiredFields
	d.Properties = properties
	return &d, nil
}

func containsRef(def Definition, targetRef string) bool {
	if def.Ref == targetRef {
		return true
	}

	for _, d := range def.Defs {
		if containsRef(d, targetRef) {
			return true
		}
	}

	for _, prop := range def.Properties {
		if containsRef(prop, targetRef) {
			return true
		}
	}

	if def.Items != nil && containsRef(*def.Items, targetRef) {
		return true
	}
	return false
}
