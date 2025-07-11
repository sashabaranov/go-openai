package jsonschema_test

import (
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

// Test Definition.Unmarshal, including success path, validation error,
// JSON syntax error and type mismatch during unmarshalling.
func TestDefinitionUnmarshal(t *testing.T) {
	schema := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"name": {Type: jsonschema.String},
		},
	}

	var dst struct {
		Name string `json:"name"`
	}
	if err := schema.Unmarshal(`{"name":"foo"}`, &dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.Name != "foo" {
		t.Errorf("expected name to be foo, got %q", dst.Name)
	}

	if err := schema.Unmarshal(`{`, &dst); err == nil {
		t.Error("expected error for malformed json")
	}

	if err := schema.Unmarshal(`{"name":1}`, &dst); err == nil {
		t.Error("expected validation error")
	}

	numSchema := jsonschema.Definition{Type: jsonschema.Number}
	var s string
	if err := numSchema.Unmarshal(`123`, &s); err == nil {
		t.Error("expected unmarshal type error")
	}
}

// Ensure GenerateSchemaForType returns an error when encountering unsupported types.
func TestGenerateSchemaForTypeUnsupported(t *testing.T) {
	type Bad struct {
		Ch chan int `json:"ch"`
	}
	if _, err := jsonschema.GenerateSchemaForType(Bad{}); err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

// Validate should fail when provided data does not match the expected container types.
func TestValidateInvalidContainers(t *testing.T) {
	objSchema := jsonschema.Definition{Type: jsonschema.Object}
	if jsonschema.Validate(objSchema, 1) {
		t.Error("expected object validation to fail for non-map input")
	}

	arrSchema := jsonschema.Definition{Type: jsonschema.Array, Items: &jsonschema.Definition{Type: jsonschema.String}}
	if jsonschema.Validate(arrSchema, 1) {
		t.Error("expected array validation to fail for non-slice input")
	}
}

// Validate should return false when $ref cannot be resolved.
func TestValidateRefNotFound(t *testing.T) {
	refSchema := jsonschema.Definition{Ref: "#/$defs/Missing"}
	if jsonschema.Validate(refSchema, "data", jsonschema.WithDefs(map[string]jsonschema.Definition{})) {
		t.Error("expected validation to fail when reference is missing")
	}
}
