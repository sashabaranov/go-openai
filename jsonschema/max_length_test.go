package jsonschema_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestDefinitionMaxLengthJSON(t *testing.T) {
	zero, three := 0, 3
	tests := []struct {
		name  string
		limit *int
		want  string
	}{
		{name: "omitted", want: `{"type":"string"}`},
		{name: "zero", limit: &zero, want: `{"type":"string","maxLength":0}`},
		{name: "positive", limit: &three, want: `{"type":"string","maxLength":3}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := jsonschema.Definition{Type: jsonschema.String, MaxLength: tt.limit}
			for _, value := range []any{schema, &schema} {
				encoded, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				if string(encoded) != tt.want {
					t.Errorf("Marshal() = %s, want %s", encoded, tt.want)
				}
				var decoded jsonschema.Definition
				if err := json.Unmarshal(encoded, &decoded); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(decoded.MaxLength, tt.limit) {
					t.Errorf("MaxLength = %v, want %v", decoded.MaxLength, tt.limit)
				}
			}
		})
	}
}

func TestFunctionDefinitionMaxLength(t *testing.T) {
	zero, three := 0, 3
	schema := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"summary": {Ref: "#/$defs/shortText"},
			"labels": {
				Type: jsonschema.Array,
				Items: &jsonschema.Definition{AnyOf: []jsonschema.Definition{
					{Type: jsonschema.String, MaxLength: &zero},
					{Type: jsonschema.Null},
				}},
			},
			"unlimited": {Type: jsonschema.String},
		},
		Defs: map[string]jsonschema.Definition{
			"shortText": {Type: jsonschema.String, MaxLength: &three},
		},
	}
	const wantJSON = `{
		"name":"record_summary",
		"parameters":{
			"type":"object",
			"properties":{
				"summary":{"$ref":"#/$defs/shortText"},
				"labels":{"type":"array","items":{"anyOf":[{"type":"string","maxLength":0},{"type":"null"}]}},
				"unlimited":{"type":"string"}
			},
			"$defs":{"shortText":{"type":"string","maxLength":3}}
		}
	}`
	var want any
	if err := json.Unmarshal([]byte(wantJSON), &want); err != nil {
		t.Fatal(err)
	}
	var wantSchema struct {
		Parameters jsonschema.Definition `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(wantJSON), &wantSchema); err != nil {
		t.Fatal(err)
	}
	for _, params := range []any{schema, &schema} {
		encoded, err := json.Marshal(openai.FunctionDefinition{Name: "record_summary", Parameters: params})
		if err != nil {
			t.Fatal(err)
		}
		var got any
		if err := json.Unmarshal(encoded, &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("function JSON = %s, want %s", encoded, wantJSON)
		}
		var decoded struct {
			Parameters jsonschema.Definition `json:"parameters"`
		}
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded.Parameters, wantSchema.Parameters) {
			t.Errorf("schema round trip = %#v, want %#v", decoded.Parameters, wantSchema.Parameters)
		}
	}
}

func TestValidateMaxLength(t *testing.T) {
	zero, one, two, three := 0, 1, 2, 3
	tests := []struct {
		name  string
		limit *int
		data  any
		enum  []string
		want  bool
	}{
		{name: "omitted", data: "unlimited", want: true},
		{name: "empty at zero", limit: &zero, data: "", want: true},
		{name: "nonempty at zero", limit: &zero, data: "a"},
		{name: "below bound", limit: &three, data: "ab", want: true},
		{name: "at bound", limit: &three, data: "abc", want: true},
		{name: "above bound", limit: &three, data: "abcd"},
		{name: "multibyte", limit: &two, data: "\u4e16\u754c", want: true},
		{name: "multibyte above bound", limit: &one, data: "\u4e16\u754c"},
		{name: "non BMP", limit: &one, data: "\U0001f600", want: true},
		{name: "combining code points", limit: &two, data: "e\u0301", want: true},
		{name: "combining above bound", limit: &one, data: "e\u0301"},
		{name: "enum match", limit: &three, data: "abc", enum: []string{"abc"}, want: true},
		{name: "enum above bound", limit: &two, data: "abc", enum: []string{"abc"}},
		{name: "enum mismatch", limit: &three, data: "ab", enum: []string{"abc"}},
		{name: "omitted with enum mismatch", data: "ab", enum: []string{"abc"}},
		{name: "wrong type", limit: &three, data: 1},
		{name: "null", limit: &three},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := jsonschema.Definition{Type: jsonschema.String, MaxLength: tt.limit, Enum: tt.enum}
			if got := jsonschema.Validate(schema, tt.data); got != tt.want {
				t.Errorf("Validate(%#v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestValidateMaxLengthComposition(t *testing.T) {
	one, three := 1, 3
	tests := []struct {
		name    string
		schema  jsonschema.Definition
		valid   any
		invalid any
	}{
		{
			name: "anyOf branch",
			schema: jsonschema.Definition{AnyOf: []jsonschema.Definition{
				{Type: jsonschema.String, MaxLength: &one}, {Type: jsonschema.Null},
			}},
			valid: nil, invalid: "ab",
		},
		{
			name: "anyOf sibling",
			schema: jsonschema.Definition{MaxLength: &one, AnyOf: []jsonschema.Definition{
				{Type: jsonschema.String}, {Type: jsonschema.Null},
			}},
			valid: "a", invalid: "ab",
		},
		{
			name: "referenced constraint",
			schema: jsonschema.Definition{Ref: "#/$defs/text", Defs: map[string]jsonschema.Definition{
				"text": {Type: jsonschema.String, MaxLength: &one},
			}},
			valid: "a", invalid: "ab",
		},
		{
			name: "ref sibling",
			schema: jsonschema.Definition{MaxLength: &one, Ref: "#/$defs/text", Defs: map[string]jsonschema.Definition{
				"text": {Type: jsonschema.String, MaxLength: &three},
			}},
			valid: "a", invalid: "ab",
		},
		{
			name:   "constraint without type",
			schema: jsonschema.Definition{MaxLength: &one},
			valid:  "a", invalid: "ab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !jsonschema.Validate(tt.schema, tt.valid) {
				t.Errorf("Validate(%v) = false, want true", tt.valid)
			}
			if jsonschema.Validate(tt.schema, tt.invalid) {
				t.Errorf("Validate(%v) = true, want false", tt.invalid)
			}
		})
	}
}

func TestMaxLengthIgnoresNonStrings(t *testing.T) {
	zero := 0
	for _, value := range []any{nil, 1, true, []any{"long"}, map[string]any{"key": "long"}} {
		if !jsonschema.Validate(jsonschema.Definition{MaxLength: &zero}, value) {
			t.Errorf("maxLength alone rejected non-string %v", value)
		}
	}
	if !jsonschema.Validate(jsonschema.Definition{Type: jsonschema.Integer, MaxLength: &zero}, 1) {
		t.Error("maxLength rejected an integer")
	}
	if jsonschema.Validate(jsonschema.Definition{}, "a") {
		t.Error("an empty schema changed its existing validation behavior")
	}
	if jsonschema.Validate(jsonschema.Definition{Ref: "#/$defs/missing", MaxLength: &zero}, "") {
		t.Error("maxLength bypassed an unresolved reference")
	}
}

func TestDefinitionUnmarshalMaxLength(t *testing.T) {
	const schemaJSON = `{
		"type":"object",
		"properties":{"names":{"type":"array","items":{"$ref":"#/$defs/name"}}},
		"$defs":{"name":{"type":"string","maxLength":1}}
	}`
	var schema jsonschema.Definition
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		t.Fatal(err)
	}
	var dst struct {
		Names []string `json:"names"`
	}
	if err := schema.Unmarshal(`{"names":["\u4e16"]}`, &dst); err != nil {
		t.Fatalf("valid Unicode name: %v", err)
	}
	if !reflect.DeepEqual(dst.Names, []string{"\u4e16"}) {
		t.Errorf("names = %v", dst.Names)
	}
	if err := schema.Unmarshal(`{"names":["ab"]}`, &dst); err == nil {
		t.Error("expected validation error for a name above maxLength")
	}
}

func ExampleDefinition_maxLength() {
	// OpenAI Structured Outputs does not support maxLength on fine-tuned models.
	maxLength := 3
	schema := jsonschema.Definition{Type: jsonschema.String, MaxLength: &maxLength}
	encoded, err := json.Marshal(&schema)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(encoded))
	fmt.Println(jsonschema.Validate(schema, "cat"))
	fmt.Println(jsonschema.Validate(schema, "cats"))
	// Output:
	// {"type":"string","maxLength":3}
	// true
	// false
}
