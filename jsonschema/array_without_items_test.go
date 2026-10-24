package jsonschema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestValidateArrayWithoutItems(t *testing.T) {
	array := jsonschema.Definition{Type: jsonschema.Array}
	strings := jsonschema.Definition{Type: jsonschema.Array, Items: &jsonschema.Definition{Type: jsonschema.String}}
	tests := []struct {
		name   string
		schema jsonschema.Definition
		data   any
		want   bool
	}{
		{"empty array", array, []any{}, true},
		{"mixed items", array, []any{1.0, "two", true, nil, map[string]any{"nested": []any{false}}}, true},
		{"string is not an array", array, "text", false},
		{"null is not an array", array, nil, false},
		{"object is not an array", array, map[string]any{}, false},
		{"number is not an array", array, 1.0, false},
		{"explicit items accepts strings", strings, []any{"one", "two"}, true},
		{"explicit items rejects numbers", strings, []any{"one", 2.0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer reportArrayValidationPanic(t)
			if got := jsonschema.Validate(tt.schema, tt.data); got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifySchemaAndUnmarshalArraysWithoutItems(t *testing.T) {
	tests := []struct {
		name    string
		schema  string
		content string
		wantErr bool
	}{
		{"direct", `{"type":"array"}`, `[1,"two",true,null,{"nested":[false]}]`, false},
		{
			"object property",
			`{"type":"object","properties":{"values":{"type":"array"}},"required":["values"]}`,
			`{"values":[1,"two"]}`, false,
		},
		{"nested arrays", `{"type":"array","items":{"type":"array"}}`, `[[1],["two",null]]`, false},
		{"reference", `{"$ref":"#/$defs/values","$defs":{"values":{"type":"array"}}}`, `["one",2]`, false},
		{"anyOf array", `{"anyOf":[{"type":"null"},{"type":"array"}]}`, `[null,false]`, false},
		{"empty array", `{"type":"array"}`, `[]`, false},
		{"anyOf null", `{"anyOf":[{"type":"null"},{"type":"array"}]}`, `null`, false},
		{"explicit item constraint", `{"type":"array","items":{"type":"string"}}`, `["one",2]`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer reportArrayValidationPanic(t)
			var schema jsonschema.Definition
			if err := json.Unmarshal([]byte(tt.schema), &schema); err != nil {
				t.Fatal(err)
			}
			var got any
			validationErr := jsonschema.VerifySchemaAndUnmarshal(schema, []byte(tt.content), &got)
			if (validationErr != nil) != tt.wantErr {
				t.Fatalf("VerifySchemaAndUnmarshal() error = %v, wantErr %v", validationErr, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			var want any
			if err := json.Unmarshal([]byte(tt.content), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("decoded value = %#v, want %#v", got, want)
			}
		})
	}
}

func reportArrayValidationPanic(t *testing.T) {
	t.Helper()
	if recovered := recover(); recovered != nil {
		t.Fatalf("array validation panicked: %v", recovered)
	}
}
