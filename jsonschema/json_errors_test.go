package jsonschema_test

import (
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

// TestGenerateSchemaForType_ErrorPaths verifies error handling for unsupported types.
func TestGenerateSchemaForType_ErrorPaths(t *testing.T) {
	type anon struct{ Ch chan int }
	tests := []struct {
		name string
		v    any
	}{
		{"slice", []chan int{}},
		{"anon struct", anon{}},
		{"pointer", (*chan int)(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := jsonschema.GenerateSchemaForType(tt.v); err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}
