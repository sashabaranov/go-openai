package jsonschema_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestValidateIntegerOutsideInt64Range(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  bool
	}{
		{"exact float64 integer boundary", math.Exp2(53), true},
		{"int64 upper boundary", math.Exp2(63), true},
		{"above int64 upper boundary", math.Nextafter(math.Exp2(63), math.Inf(1)), true},
		{"int64 lower boundary", -math.Exp2(63), true},
		{"below int64 lower boundary", math.Nextafter(-math.Exp2(63), math.Inf(-1)), true},
		{"large positive integer", 1e100, true},
		{"large negative integer", -1e100, true},
		{"positive fraction", 1.5, false},
		{"negative fraction", -1.5, false},
		{"positive infinity", math.Inf(1), false},
		{"negative infinity", math.Inf(-1), false},
		{"not a number", math.NaN(), false},
		{"zero", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonschema.Validate(jsonschema.Definition{Type: jsonschema.Integer}, tt.value); got != tt.want {
				t.Errorf("Validate(integer, %v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestUnmarshalLargeUnsignedInteger(t *testing.T) {
	type result struct {
		ID uint64 `json:"id"`
	}
	schema, err := jsonschema.GenerateSchemaForType(result{})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []uint64{1, 1<<53 + 1, math.MaxInt64, 1 << 63, math.MaxUint64} {
		want := result{ID: id}
		content, marshalErr := json.Marshal(want)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		var got result
		if unmarshalErr := schema.Unmarshal(string(content), &got); unmarshalErr != nil {
			t.Errorf("Unmarshal(%s): %v", content, unmarshalErr)
		} else if got != want {
			t.Errorf("Unmarshal(%s) = %+v, want %+v", content, got, want)
		}
	}
	for _, content := range []string{`{"id":18446744073709551616}`, `{"id":-1}`, `{"id":1.5}`} {
		var got result
		if unmarshalErr := schema.Unmarshal(content, &got); unmarshalErr == nil {
			t.Errorf("Unmarshal(%s) unexpectedly succeeded", content)
		}
	}
}
