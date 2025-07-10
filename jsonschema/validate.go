package jsonschema

import (
	"encoding/json"
	"errors"
)

func CollectDefs(def Definition) map[string]Definition {
	result := make(map[string]Definition)
	collectDefsRecursive(def, result, "#")
	return result
}

func collectDefsRecursive(def Definition, result map[string]Definition, prefix string) {
	for k, v := range def.Defs {
		path := prefix + "/$defs/" + k
		result[path] = v
		collectDefsRecursive(v, result, path)
	}
	for k, sub := range def.Properties {
		collectDefsRecursive(sub, result, prefix+"/properties/"+k)
	}
	if def.Items != nil {
		collectDefsRecursive(*def.Items, result, prefix)
	}
}

func VerifySchemaAndUnmarshal(schema Definition, content []byte, v any) error {
	var data any
	err := json.Unmarshal(content, &data)
	if err != nil {
		return err
	}
	if !Validate(schema, data, WithDefs(CollectDefs(schema))) {
		return errors.New("data validation failed against the provided schema")
	}
	return json.Unmarshal(content, &v)
}

type validateArgs struct {
	Defs map[string]Definition
}

type ValidateOption func(*validateArgs)

func WithDefs(defs map[string]Definition) ValidateOption {
	return func(option *validateArgs) {
		option.Defs = defs
	}
}

func Validate(schema Definition, data any, opts ...ValidateOption) bool {
	args := validateArgs{}
	for _, opt := range opts {
		opt(&args)
	}
	if len(opts) == 0 {
		args.Defs = CollectDefs(schema)
	}
	switch schema.Type {
	case Object:
		return validateObject(schema, data, args.Defs)
	case Array:
		return validateArray(schema, data, args.Defs)
	case String:
		v, ok := data.(string)
		if ok && len(schema.Enum) > 0 {
			return contains(schema.Enum, v)
		}
		return ok
	case Number: // float64 and int
		_, ok := data.(float64)
		if !ok {
			_, ok = data.(int)
		}
		return ok
	case Boolean:
		_, ok := data.(bool)
		return ok
	case Integer:
		// Golang unmarshals all numbers as float64, so we need to check if the float64 is an integer
		if num, ok := data.(float64); ok {
			return num == float64(int64(num))
		}
		_, ok := data.(int)
		return ok
	case Null:
		return data == nil
	default:
		if schema.Ref != "" && args.Defs != nil {
			if v, ok := args.Defs[schema.Ref]; ok {
				return Validate(v, data, WithDefs(args.Defs))
			}
		}
		return false
	}
}

func validateObject(schema Definition, data any, defs map[string]Definition) bool {
	dataMap, ok := data.(map[string]any)
	if !ok {
		return false
	}
	for _, field := range schema.Required {
		if _, exists := dataMap[field]; !exists {
			return false
		}
	}
	for key, valueSchema := range schema.Properties {
		value, exists := dataMap[key]
		if exists && !Validate(valueSchema, value, WithDefs(defs)) {
			return false
		} else if !exists && contains(schema.Required, key) {
			return false
		}
	}
	return true
}

func validateArray(schema Definition, data any, defs map[string]Definition) bool {
	dataArray, ok := data.([]any)
	if !ok {
		return false
	}
	for _, item := range dataArray {
		if !Validate(*schema.Items, item, WithDefs(defs)) {
			return false
		}
	}
	return true
}

func contains[S ~[]E, E comparable](s S, v E) bool {
	for i := range s {
		if v == s[i] {
			return true
		}
	}
	return false
}
