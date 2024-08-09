package jsonschema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestDefinition_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		def  jsonschema.Definition
		want string
	}{
		{
			name: "Test with empty Definition",
			def:  jsonschema.Definition{},
			want: `{"properties":{}}`,
		},
		{
			name: "Test with Definition properties set",
			def: jsonschema.Definition{
				Type:        jsonschema.String,
				Description: "A string type",
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: jsonschema.String,
					},
				},
			},
			want: `{
   "type":"string",
   "description":"A string type",
   "properties":{
      "name":{
         "type":"string",
         "properties":{}
      }
   }
}`,
		},
		{
			name: "Test with nested Definition properties",
			def: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"user": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"name": {
								Type: jsonschema.String,
							},
							"age": {
								Type: jsonschema.Integer,
							},
						},
					},
				},
			},
			want: `{
   "type":"object",
   "properties":{
      "user":{
         "type":"object",
         "properties":{
            "name":{
               "type":"string",
               "properties":{}
            },
            "age":{
               "type":"integer",
               "properties":{}
            }
         }
      }
   }
}`,
		},
		{
			name: "Test with complex nested Definition",
			def: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"user": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"name": {
								Type: jsonschema.String,
							},
							"age": {
								Type: jsonschema.Integer,
							},
							"address": {
								Type: jsonschema.Object,
								Properties: map[string]jsonschema.Definition{
									"city": {
										Type: jsonschema.String,
									},
									"country": {
										Type: jsonschema.String,
									},
								},
							},
						},
					},
				},
			},
			want: `{
   "type":"object",
   "properties":{
      "user":{
         "type":"object",
         "properties":{
            "name":{
               "type":"string",
               "properties":{}
            },
            "age":{
               "type":"integer",
               "properties":{}
            },
            "address":{
               "type":"object",
               "properties":{
                  "city":{
                     "type":"string",
                     "properties":{}
                  },
                  "country":{
                     "type":"string",
                     "properties":{}
                  }
               }
            }
         }
      }
   }
}`,
		},
		{
			name: "Test with Array type Definition",
			def: jsonschema.Definition{
				Type: jsonschema.Array,
				Items: &jsonschema.Definition{
					Type: jsonschema.String,
				},
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: jsonschema.String,
					},
				},
			},
			want: `{
   "type":"array",
   "items":{
      "type":"string",
      "properties":{
         
      }
   },
   "properties":{
      "name":{
         "type":"string",
         "properties":{}
      }
   }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantBytes := []byte(tt.want)
			var want map[string]interface{}
			err := json.Unmarshal(wantBytes, &want)
			if err != nil {
				t.Errorf("Failed to Unmarshal JSON: error = %v", err)
				return
			}

			got := structToMap(t, tt.def)
			gotPtr := structToMap(t, &tt.def)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, want)
			}
			if !reflect.DeepEqual(gotPtr, want) {
				t.Errorf("MarshalJSON() gotPtr = %v, want %v", gotPtr, want)
			}
		})
	}
}

func structToMap(t *testing.T, v any) map[string]any {
	t.Helper()
	gotBytes, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Failed to Marshal JSON: error = %v", err)
		return nil
	}

	var got map[string]interface{}
	err = json.Unmarshal(gotBytes, &got)
	if err != nil {
		t.Errorf("Failed to Unmarshal JSON: error =  %v", err)
		return nil
	}
	return got
}

type MyStructuredResponse struct {
	PascalCase string   `json:"pascal_case,omitempty" required:"true" description:"PascalCase"`
	CamelCase  string   `json:"camel_case" required:"true" description:"CamelCase"`
	KebabCase  string   `json:"kebab_case,omitempty" required:"false" description:"KebabCase"`
	SnakeCase  string   `json:"snake_case" required:"true" description:"SnakeCase"`
	Keywords   []string `json:"keywords,omitempty" description:"Keywords"`
	Optional   bool     `json:"optional,omitempty"`
}

func TestWrap(t *testing.T) {
	schemaStr := `{
  "type": "object",
  "properties": {
    "camel_case": {
      "type": "string",
      "description": "CamelCase"
    },
    "kebab_case": {
      "type": "string",
      "description": "KebabCase"
    },
    "keywords": {
      "type": "array",
      "description": "Keywords",
      "items": {
        "type": "string"
      }
    },
    "optional": {
      "type": "boolean"
    },
    "pascal_case": {
      "type": "string",
      "description": "PascalCase"
    },
    "snake_case": {
      "type": "string",
      "description": "SnakeCase"
    }
  },
  "required": [
    "pascal_case",
    "camel_case",
    "snake_case"
  ],
  "additionalProperties": false
}`
	schema, err := jsonschema.Wrap(MyStructuredResponse{})
	if err != nil {
		t.Fatal(err)
	}
	if schema.String() != schemaStr {
		t.Errorf("Failed to Generate JSONSchema: schema =  %s", schema)
	}
	type CustomStruct struct {
		Title   string                `json:"title"`
		Data    *MyStructuredResponse `json:"data,omitempty"`
		private string
	}
	schema2Str := `{
  "type": "object",
  "properties": {
    "data": {
      "type": "object",
      "properties": {
        "camel_case": {
          "type": "string",
          "description": "CamelCase"
        },
        "kebab_case": {
          "type": "string",
          "description": "KebabCase"
        },
        "keywords": {
          "type": "array",
          "description": "Keywords",
          "items": {
            "type": "string"
          }
        },
        "optional": {
          "type": "boolean"
        },
        "pascal_case": {
          "type": "string",
          "description": "PascalCase"
        },
        "snake_case": {
          "type": "string",
          "description": "SnakeCase"
        }
      },
      "required": [
        "pascal_case",
        "camel_case",
        "snake_case"
      ],
      "additionalProperties": false
    },
    "title": {
      "type": "string"
    }
  },
  "required": [
    "title"
  ],
  "additionalProperties": false
}`
	schema2, err := jsonschema.Wrap(CustomStruct{})
	if err != nil {
		t.Fatal(err)
	}
	if schema2.String() != schema2Str {
		t.Errorf("Failed to Generate JSONSchema: schema =  %s", schema)
	}
}

func TestSchemaWrapper_Unmarshal(t *testing.T) {
	schema, err := jsonschema.Wrap(MyStructuredResponse{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := schema.Unmarshal(`{"pascal_case":"a","camel_case":"b","snake_case":"c","keywords":[]}`)
	if err != nil {
		t.Errorf("Failed to SchemaWrapper Unmarshal: error =  %v", err)
	} else {
		var v = MyStructuredResponse{
			PascalCase: "a",
			CamelCase:  "b",
			SnakeCase:  "c",
			Keywords:   []string{},
		}
		if !reflect.DeepEqual(*result, v) {
			t.Errorf("Failed to SchemaWrapper Unmarshal: result =  %v", *result)
		}
	}
}
