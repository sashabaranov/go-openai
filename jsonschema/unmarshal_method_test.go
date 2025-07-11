package jsonschema_test

import (
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestDefinition_Unmarshal(t *testing.T) {
	schema := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"name": {Type: jsonschema.String},
			"age":  {Type: jsonschema.Integer},
		},
		Required: []string{"name", "age"},
	}
	var out struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	err := schema.Unmarshal(`{"name":"Alice","age":25}`, &out)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.Name != "Alice" || out.Age != 25 {
		t.Fatalf("unexpected result %+v", out)
	}

	if err2 := schema.Unmarshal(`{"name":"Bob"}`, &out); err2 == nil {
		t.Fatalf("expected error for missing field")
	}
}
