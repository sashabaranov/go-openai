package jsonschema

import (
	"testing"
)

func Test_Validate(t *testing.T) {
	type args struct {
		data   interface{}
		schema Definition
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// string integer number boolean
		{"", args{data: "ABC", schema: Definition{Type: String}}, true},
		{"", args{data: 123, schema: Definition{Type: String}}, false},
		{"", args{data: 123, schema: Definition{Type: Integer}}, true},
		{"", args{data: 123.4, schema: Definition{Type: Integer}}, false},
		{"", args{data: "ABC", schema: Definition{Type: Number}}, false},
		{"", args{data: 123, schema: Definition{Type: Number}}, true},
		{"", args{data: false, schema: Definition{Type: Boolean}}, true},
		{"", args{data: 123, schema: Definition{Type: Boolean}}, false},
		{"", args{data: nil, schema: Definition{Type: Null}}, true},
		{"", args{data: 0, schema: Definition{Type: Null}}, false},
		// array
		{"", args{data: []any{"a", "b", "c"}, schema: Definition{Type: Array, Items: &Definition{Type: String}}}, true},
		{"", args{data: []any{1, 2, 3}, schema: Definition{Type: Array, Items: &Definition{Type: String}}}, false},
		{"", args{data: []any{1, 2, 3}, schema: Definition{Type: Array, Items: &Definition{Type: Integer}}}, true},
		{"", args{data: []any{1, 2, 3.4}, schema: Definition{Type: Array, Items: &Definition{Type: Integer}}}, false},
		// object
		{"", args{data: map[string]any{
			"string":  "abc",
			"integer": 123,
			"number":  123.4,
			"boolean": false,
			"array":   []any{1, 2, 3},
		}, schema: Definition{Type: Object, Properties: map[string]Definition{
			"string":  {Type: String},
			"integer": {Type: Integer},
			"number":  {Type: Number},
			"boolean": {Type: Boolean},
			"array":   {Type: Array, Items: &Definition{Type: Number}},
		},
			Required: []string{"string"},
		}}, true},
		{"", args{data: map[string]any{
			"integer": 123,
			"number":  123.4,
			"boolean": false,
			"array":   []any{1, 2, 3},
		}, schema: Definition{Type: Object, Properties: map[string]Definition{
			"string":  {Type: String},
			"integer": {Type: Integer},
			"number":  {Type: Number},
			"boolean": {Type: Boolean},
			"array":   {Type: Array, Items: &Definition{Type: Number}},
		},
			Required: []string{"string"},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Validate(tt.args.schema, tt.args.data); got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type args struct {
		schema  Definition
		content []byte
		v       any
	}
	var result1 struct {
		String string  `json:"string"`
		Number float64 `json:"number"`
	}
	var result2 struct {
		String string  `json:"string"`
		Number float64 `json:"number"`
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{
			schema: Definition{
				Type: Object,
				Properties: map[string]Definition{
					"string": {Type: String},
					"number": {Type: Number},
				},
			},
			content: []byte(`{"string":"abc","number":123.4}`),
			v:       &result1,
		}, false},
		{"", args{
			schema: Definition{
				Type: Object,
				Properties: map[string]Definition{
					"string": {Type: String},
					"number": {Type: Number},
				},
				Required: []string{"string", "number"},
			},
			content: []byte(`{"string":"abc"}`),
			v:       result2,
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unmarshal(tt.args.schema, tt.args.content, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				t.Logf("Unmarshal() v = %+v\n", tt.args.v)
			}
		})
	}
}
