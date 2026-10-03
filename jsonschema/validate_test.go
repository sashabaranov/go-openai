package jsonschema_test

import (
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestValidateAdditionalProperties(t *testing.T) {
	properties := map[string]jsonschema.Definition{"name": {Type: jsonschema.String}}
	valid := map[string]any{"name": "test"}
	extra := map[string]any{"name": "test", "extra": true}
	invalid := map[string]any{"name": false}
	tests := []struct {
		name                 string
		properties           map[string]jsonschema.Definition
		additionalProperties any
		data                 map[string]any
		want                 bool
	}{
		{"false allows declared properties", properties, false, valid, true},
		{"false rejects extra properties", properties, false, extra, false},
		{"true allows extra properties", properties, true, extra, true},
		{"unset allows extra properties", properties, nil, extra, true},
		{"empty properties allows empty object", map[string]jsonschema.Definition{}, false, map[string]any{}, true},
		{"empty properties rejects extra properties", map[string]jsonschema.Definition{}, false, extra, false},
		{"nil properties allows empty object", nil, false, map[string]any{}, true},
		{"nil properties rejects extra properties", nil, false, extra, false},
		{"true still validates declared properties", properties, true, invalid, false},
		{"unset still validates declared properties", properties, nil, invalid, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := jsonschema.Definition{
				Type:                 jsonschema.Object,
				Properties:           tt.properties,
				AdditionalProperties: tt.additionalProperties,
			}
			if got := jsonschema.Validate(schema, tt.data); got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAdditionalPropertiesRecursive(t *testing.T) {
	closed := jsonschema.Definition{
		Type:                 jsonschema.Object,
		Properties:           map[string]jsonschema.Definition{"name": {Type: jsonschema.String}},
		Required:             []string{"name"},
		AdditionalProperties: false,
	}
	nested := jsonschema.Definition{
		Type:       jsonschema.Object,
		Properties: map[string]jsonschema.Definition{"child": closed},
	}
	array := jsonschema.Definition{Type: jsonschema.Array, Items: &closed}
	ref := jsonschema.Definition{
		Ref:  "#/$defs/Closed",
		Defs: map[string]jsonschema.Definition{"Closed": closed},
	}
	union := jsonschema.Definition{AnyOf: []jsonschema.Definition{closed, {Type: jsonschema.String}}}
	valid := map[string]any{"name": "test"}
	extra := map[string]any{"name": "test", "extra": true}
	tests := []struct {
		name   string
		schema jsonschema.Definition
		data   any
		want   bool
	}{
		{"required property still required", closed, map[string]any{}, false},
		{"nested valid object", nested, map[string]any{"child": valid}, true},
		{"nested extra property", nested, map[string]any{"child": extra}, false},
		{"outer extra property still allowed", nested, map[string]any{"child": valid, "extra": true}, true},
		{"array valid objects", array, []any{valid, valid}, true},
		{"array extra property", array, []any{valid, extra}, false},
		{"ref valid object", ref, valid, true},
		{"ref extra property", ref, extra, false},
		{"anyOf valid object", union, valid, true},
		{"anyOf other type", union, "test", true},
		{"anyOf extra property", union, extra, false},
		{"anyOf another object allows extras", jsonschema.Definition{
			AnyOf: []jsonschema.Definition{closed, {Type: jsonschema.Object}},
		}, extra, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonschema.Validate(tt.schema, tt.data); got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratedSchemaRejectsAdditionalProperties(t *testing.T) {
	type result struct {
		Name string `json:"name"`
	}
	schema, err := jsonschema.GenerateSchemaForType(result{})
	if err != nil {
		t.Fatal(err)
	}
	unmarshalers := []struct {
		name      string
		unmarshal func(string, any) error
	}{
		{"Definition.Unmarshal", schema.Unmarshal},
		{"VerifySchemaAndUnmarshal", func(content string, v any) error {
			return jsonschema.VerifySchemaAndUnmarshal(*schema, []byte(content), v)
		}},
	}
	for _, unmarshal := range unmarshalers {
		t.Run(unmarshal.name, func(t *testing.T) {
			got := result{Name: "original"}
			if err := unmarshal.unmarshal(`{"name":"test","extra":true}`, &got); err == nil {
				t.Fatal("expected validation error for extra property")
			}
			if got.Name != "original" {
				t.Errorf("rejected data changed destination to %+v", got)
			}
			if err := unmarshal.unmarshal(`{"name":"test"}`, &got); err != nil {
				t.Fatal(err)
			}
			if got.Name != "test" {
				t.Errorf("Name = %q, want test", got.Name)
			}
		})
	}
}

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
		// anyOf: nullable string union
		{"", args{data: "abc", schema: jsonschema.Definition{AnyOf: []jsonschema.Definition{
			{Type: jsonschema.String}, {Type: jsonschema.Null},
		}}}, true},
		{"", args{data: nil, schema: jsonschema.Definition{AnyOf: []jsonschema.Definition{
			{Type: jsonschema.String}, {Type: jsonschema.Null},
		}}}, true},
		{"", args{data: 123, schema: jsonschema.Definition{AnyOf: []jsonschema.Definition{
			{Type: jsonschema.String}, {Type: jsonschema.Null},
		}}}, false},
		// anyOf nested in an object property
		{"", args{data: map[string]any{"summary": nil}, schema: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"summary": {AnyOf: []jsonschema.Definition{{Type: jsonschema.String}, {Type: jsonschema.Null}}},
			},
			Required: []string{"summary"},
		}}, true},
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
