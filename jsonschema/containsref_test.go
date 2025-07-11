package jsonschema_test

import (
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

// SelfRef struct used to produce a self-referential schema.
type SelfRef struct {
	Friends []SelfRef `json:"friends"`
}

// Address struct referenced by Person without self-reference.
type Address struct {
	Street string `json:"street"`
}

type Person struct {
	Address Address `json:"address"`
}

// TestGenerateSchemaForType_SelfRef ensures that self-referential types are not
// flattened during schema generation.
func TestGenerateSchemaForType_SelfRef(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(SelfRef{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := schema.Defs["SelfRef"]; !ok {
		t.Fatal("expected defs to contain SelfRef for self reference")
	}
}

// TestGenerateSchemaForType_NoSelfRef ensures that non-self-referential types
// are flattened and do not reappear in $defs.
func TestGenerateSchemaForType_NoSelfRef(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(Person{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := schema.Defs["Person"]; ok {
		t.Fatal("unexpected Person definition in defs")
	}
	if _, ok := schema.Defs["Address"]; !ok {
		t.Fatal("expected Address definition in defs")
	}
}
