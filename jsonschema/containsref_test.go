package jsonschema

import "testing"

// TestContainsRef ensures containsRef recursively searches definitions and items.
func TestContainsRef(t *testing.T) {
	schema := Definition{
		Type: Object,
		Properties: map[string]Definition{
			"person": {Ref: "#/$defs/Person"},
		},
		Defs: map[string]Definition{
			"Person": {
				Type: Object,
				Properties: map[string]Definition{
					"friends": {
						Type:  Array,
						Items: &Definition{Ref: "#/$defs/Person"},
					},
				},
			},
		},
	}

	if !containsRef(schema, "#/$defs/Person") {
		t.Fatal("expected to find reference in root")
	}
	if !containsRef(schema.Defs["Person"], "#/$defs/Person") {
		t.Fatal("expected to find self reference in defs")
	}
	if containsRef(schema, "#/$defs/Unknown") {
		t.Fatal("unexpected reference found")
	}
}
