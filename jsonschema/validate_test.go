package jsonschema_test

import (
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func Test_Validate(t *testing.T) {
	type args struct {
		data   any
		schema jsonschema.Definition
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// string integer number boolean
		{"", args{data: "ABC", schema: jsonschema.Definition{Type: jsonschema.String}}, true},
		{"", args{data: 123, schema: jsonschema.Definition{Type: jsonschema.String}}, false},
		{"", args{data: 123, schema: jsonschema.Definition{Type: jsonschema.Integer}}, true},
		{"", args{data: 123.4, schema: jsonschema.Definition{Type: jsonschema.Integer}}, false},
		{"", args{data: "ABC", schema: jsonschema.Definition{Type: jsonschema.Number}}, false},
		{"", args{data: 123, schema: jsonschema.Definition{Type: jsonschema.Number}}, true},
		{"", args{data: false, schema: jsonschema.Definition{Type: jsonschema.Boolean}}, true},
		{"", args{data: 123, schema: jsonschema.Definition{Type: jsonschema.Boolean}}, false},
		{"", args{data: nil, schema: jsonschema.Definition{Type: jsonschema.Null}}, true},
		{"", args{data: 0, schema: jsonschema.Definition{Type: jsonschema.Null}}, false},
		// array
		{"", args{data: []any{"a", "b", "c"}, schema: jsonschema.Definition{
			Type: jsonschema.Array, Items: &jsonschema.Definition{Type: jsonschema.String}},
		}, true},
		{"", args{data: []any{1, 2, 3}, schema: jsonschema.Definition{
			Type: jsonschema.Array, Items: &jsonschema.Definition{Type: jsonschema.String}},
		}, false},
		{"", args{data: []any{1, 2, 3}, schema: jsonschema.Definition{
			Type: jsonschema.Array, Items: &jsonschema.Definition{Type: jsonschema.Integer}},
		}, true},
		{"", args{data: []any{1, 2, 3.4}, schema: jsonschema.Definition{
			Type: jsonschema.Array, Items: &jsonschema.Definition{Type: jsonschema.Integer}},
		}, false},
		// object
		{"", args{data: map[string]any{
			"string":  "abc",
			"integer": 123,
			"number":  123.4,
			"boolean": false,
			"array":   []any{1, 2, 3},
		}, schema: jsonschema.Definition{Type: jsonschema.Object, Properties: map[string]jsonschema.Definition{
			"string":  {Type: jsonschema.String},
			"integer": {Type: jsonschema.Integer},
			"number":  {Type: jsonschema.Number},
			"boolean": {Type: jsonschema.Boolean},
			"array":   {Type: jsonschema.Array, Items: &jsonschema.Definition{Type: jsonschema.Number}},
		},
			Required: []string{"string"},
		}}, true},
		{"", args{data: map[string]any{
			"integer": 123,
			"number":  123.4,
			"boolean": false,
			"array":   []any{1, 2, 3},
		}, schema: jsonschema.Definition{Type: jsonschema.Object, Properties: map[string]jsonschema.Definition{
			"string":  {Type: jsonschema.String},
			"integer": {Type: jsonschema.Integer},
			"number":  {Type: jsonschema.Number},
			"boolean": {Type: jsonschema.Boolean},
			"array":   {Type: jsonschema.Array, Items: &jsonschema.Definition{Type: jsonschema.Number}},
		},
			Required: []string{"string"},
		}}, false},
		{
			"test schema with ref and defs", args{data: map[string]any{
				"person": map[string]any{
					"name":   "John",
					"gender": "male",
					"age":    28,
					"profile": map[string]any{
						"full_name": "John Doe",
					},
				},
			}, schema: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"person": {Ref: "#/$defs/Person"},
				},
				Required: []string{"person"},
				Defs: map[string]jsonschema.Definition{
					"Person": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"name":    {Type: jsonschema.String},
							"gender":  {Type: jsonschema.String, Enum: []string{"male", "female", "unknown"}},
							"age":     {Type: jsonschema.Integer},
							"profile": {Ref: "#/$defs/Person/$defs/Profile"},
							"tweets":  {Type: jsonschema.Array, Items: &jsonschema.Definition{Ref: "#/$defs/Tweet"}},
						},
						Required: []string{"name", "gender", "age", "profile"},
						Defs: map[string]jsonschema.Definition{
							"Profile": {
								Type: jsonschema.Object,
								Properties: map[string]jsonschema.Definition{
									"full_name": {Type: jsonschema.String},
								},
							},
						},
					},
					"Tweet": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"text":   {Type: jsonschema.String},
							"person": {Ref: "#/$defs/Person"},
						},
					},
				},
			}}, true},
		{
			"test enum invalid value", args{data: map[string]any{
				"person": map[string]any{
					"name":   "John",
					"gender": "other",
					"age":    28,
					"profile": map[string]any{
						"full_name": "John Doe",
					},
				},
			}, schema: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"person": {Ref: "#/$defs/Person"},
				},
				Required: []string{"person"},
				Defs: map[string]jsonschema.Definition{
					"Person": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"name":    {Type: jsonschema.String},
							"gender":  {Type: jsonschema.String, Enum: []string{"male", "female", "unknown"}},
							"age":     {Type: jsonschema.Integer},
							"profile": {Ref: "#/$defs/Person/$defs/Profile"},
							"tweets":  {Type: jsonschema.Array, Items: &jsonschema.Definition{Ref: "#/$defs/Tweet"}},
						},
						Required: []string{"name", "gender", "age", "profile"},
						Defs: map[string]jsonschema.Definition{
							"Profile": {
								Type: jsonschema.Object,
								Properties: map[string]jsonschema.Definition{
									"full_name": {Type: jsonschema.String},
								},
							},
						},
					},
					"Tweet": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"text":   {Type: jsonschema.String},
							"person": {Ref: "#/$defs/Person"},
						},
					},
				},
			}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonschema.Validate(tt.args.schema, tt.args.data); got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type args struct {
		schema  jsonschema.Definition
		content []byte
		v       any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{
			schema: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"string": {Type: jsonschema.String},
					"number": {Type: jsonschema.Number},
				},
			},
			content: []byte(`{"string":"abc","number":123.4}`),
			v: &struct {
				String string  `json:"string"`
				Number float64 `json:"number"`
			}{},
		}, false},
		{"", args{
			schema: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"string": {Type: jsonschema.String},
					"number": {Type: jsonschema.Number},
				},
				Required: []string{"string", "number"},
			},
			content: []byte(`{"string":"abc"}`),
			v: struct {
				String string  `json:"string"`
				Number float64 `json:"number"`
			}{},
		}, true},
		{"validate integer", args{
			schema: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"string":  {Type: jsonschema.String},
					"integer": {Type: jsonschema.Integer},
				},
				Required: []string{"string", "integer"},
			},
			content: []byte(`{"string":"abc","integer":123}`),
			v: &struct {
				String  string `json:"string"`
				Integer int    `json:"integer"`
			}{},
		}, false},
		{"validate integer failed", args{
			schema: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"string":  {Type: jsonschema.String},
					"integer": {Type: jsonschema.Integer},
				},
				Required: []string{"string", "integer"},
			},
			content: []byte(`{"string":"abc","integer":123.4}`),
			v: &struct {
				String  string `json:"string"`
				Integer int    `json:"integer"`
			}{},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jsonschema.VerifySchemaAndUnmarshal(tt.args.schema, tt.args.content, tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollectDefs(t *testing.T) {
	type args struct {
		schema jsonschema.Definition
	}
	tests := []struct {
		name string
		args args
		want map[string]jsonschema.Definition
	}{
		{
			"test collect defs",
			args{
				schema: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"person": {Ref: "#/$defs/Person"},
					},
					Required: []string{"person"},
					Defs: map[string]jsonschema.Definition{
						"Person": {
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"name":    {Type: jsonschema.String},
								"gender":  {Type: jsonschema.String, Enum: []string{"male", "female", "unknown"}},
								"age":     {Type: jsonschema.Integer},
								"profile": {Ref: "#/$defs/Person/$defs/Profile"},
								"tweets":  {Type: jsonschema.Array, Items: &jsonschema.Definition{Ref: "#/$defs/Tweet"}},
							},
							Required: []string{"name", "gender", "age", "profile"},
							Defs: map[string]jsonschema.Definition{
								"Profile": {
									Type: jsonschema.Object,
									Properties: map[string]jsonschema.Definition{
										"full_name": {Type: jsonschema.String},
									},
								},
							},
						},
						"Tweet": {
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"text":   {Type: jsonschema.String},
								"person": {Ref: "#/$defs/Person"},
							},
						},
					},
				},
			},
			map[string]jsonschema.Definition{
				"#/$defs/Person": {
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"name":    {Type: jsonschema.String},
						"gender":  {Type: jsonschema.String, Enum: []string{"male", "female", "unknown"}},
						"age":     {Type: jsonschema.Integer},
						"profile": {Ref: "#/$defs/Person/$defs/Profile"},
						"tweets":  {Type: jsonschema.Array, Items: &jsonschema.Definition{Ref: "#/$defs/Tweet"}},
					},
					Required: []string{"name", "gender", "age", "profile"},
					Defs: map[string]jsonschema.Definition{
						"Profile": {
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"full_name": {Type: jsonschema.String},
							},
						},
					},
				},
				"#/$defs/Person/$defs/Profile": {
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"full_name": {Type: jsonschema.String},
					},
				},
				"#/$defs/Tweet": {
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"text":   {Type: jsonschema.String},
						"person": {Ref: "#/$defs/Person"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jsonschema.CollectDefs(tt.args.schema)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CollectDefs() = %v, want %v", got, tt.want)
			}
		})
	}
}
