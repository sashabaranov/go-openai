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
			want: `{}`,
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
							"type":"string"
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
										"type":"string"
									},
									"age":{
										"type":"integer"
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
										"type":"string"
									},
									"age":{
										"type":"integer"
									},
									"address":{
										"type":"object",
										"properties":{
												"city":{
													"type":"string"
												},
												"country":{
													"type":"string"
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
						"type":"string"
				},
				"properties":{
						"name":{
							"type":"string"
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

type User struct {
	ID     int      `json:"id,omitempty"`
	Name   string   `json:"name,omitempty"`
	Orders []*Order `json:"orders,omitempty"`
}

type Order struct {
	ID     int     `json:"id,omitempty"`
	Amount float64 `json:"amount,omitempty"`
	Buyer  *User   `json:"buyer,omitempty"`
}

func TestStructToSchema(t *testing.T) {
	type Tweet struct {
		Text string `json:"text"`
	}

	type Person struct {
		Name    string   `json:"name,omitempty"`
		Age     int      `json:"age,omitempty"`
		Friends []Person `json:"friends,omitempty"`
		Tweets  []Tweet  `json:"tweets,omitempty"`
	}

	type MyStructuredResponse struct {
		PascalCase string `json:"pascal_case" required:"true" description:"PascalCase"`
		CamelCase  string `json:"camel_case" required:"true" description:"CamelCase"`
		KebabCase  string `json:"kebab_case" required:"true" description:"KebabCase"`
		SnakeCase  string `json:"snake_case" required:"true" description:"SnakeCase"`
	}

	tests := []struct {
		name string
		in   any
		want string
	}{
		{
			name: "Test with empty struct",
			in:   struct{}{},
			want: `{
				"type":"object",
				"additionalProperties":false
			}`,
		},
		{
			name: "Test with struct containing many fields",
			in: struct {
				Name   string  `json:"name"`
				Age    int     `json:"age"`
				Active bool    `json:"active"`
				Height float64 `json:"height"`
				Cities []struct {
					Name  string `json:"name"`
					State string `json:"state"`
				} `json:"cities"`
			}{
				Name: "John Doe",
				Age:  30,
				Cities: []struct {
					Name  string `json:"name"`
					State string `json:"state"`
				}{
					{Name: "New York", State: "NY"},
					{Name: "Los Angeles", State: "CA"},
				},
			},
			want: `{
				"type":"object",
				"properties":{
					"name":{
						"type":"string"
					},
					"age":{
						"type":"integer"
					},
					"active":{
						"type":"boolean"
					},
					"height":{
						"type":"number"
					},
					"cities":{
						"type":"array",
						"items":{
							"additionalProperties":false,
							"type":"object",
							"properties":{
								"name":{
									"type":"string"
								},
								"state":{
									"type":"string"
								}
							},
							"required":["name","state"]
						}
					}
				},
				"required":["name","age","active","height","cities"],
				"additionalProperties":false
			}`,
		},
		{
			name: "Test with description tag",
			in: struct {
				Name string `json:"name" description:"The name of the person"`
			}{
				Name: "John Doe",
			},
			want: `{
				"type":"object",
				"properties":{
					"name":{
						"type":"string",
						"description":"The name of the person"
					}
				},
				"required":["name"],
				"additionalProperties":false
			}`,
		},
		{
			name: "Test with required tag",
			in: struct {
				Name string `json:"name" required:"false"`
			}{
				Name: "John Doe",
			},
			want: `{
				"type":"object",
				"properties":{
					"name":{
						"type":"string"
					}
				},
				"additionalProperties":false
			}`,
		},
		{
			name: "Test with enum tag",
			in: struct {
				Color string `json:"color" enum:"red,green,blue"`
			}{
				Color: "red",
			},
			want: `{
				"type":"object",
				"properties":{
					"color":{
						"type":"string",
						"enum":["red","green","blue"]
					}
				},
				"required":["color"],
				"additionalProperties":false
			}`,
		},
		{
			name: "Test with nullable tag",
			in: struct {
				Name *string `json:"name" nullable:"true"`
			}{
				Name: nil,
			},
			want: `{

				"type":"object",
				"properties":{
					"name":{
						"type":"string",
						"nullable":true
					}
				},
				"required":["name"],
				"additionalProperties":false
			}`,
		},
		{
			name: "Test with exclude mark",
			in: struct {
				Name string `json:"-"`
			}{
				Name: "Name",
			},
			want: `{
				"type":"object",
				"additionalProperties":false
			}`,
		},
		{
			name: "Test with no json tag",
			in: struct {
				Name string
			}{
				Name: "",
			},
			want: `{
				"type":"object",
				"properties":{
					"Name":{
						"type":"string"
					}
				},
				"required":["Name"],
				"additionalProperties":false
			}`,
		},
		{
			name: "Test with omitempty tag",
			in: struct {
				Name string `json:"name,omitempty"`
			}{
				Name: "",
			},
			want: `{
				"type":"object",
				"properties":{
					"name":{
						"type":"string"
					}
				},
				"additionalProperties":false
			}`,
		},
		{
			name: "Test with $ref and $defs",
			in: struct {
				Person Person  `json:"person"`
				Tweets []Tweet `json:"tweets"`
			}{},
			want: `{
  "type" : "object",
  "properties" : {
    "person" : {
      "$ref" : "#/$defs/Person"
    },
    "tweets" : {
      "type" : "array",
      "items" : {
        "$ref" : "#/$defs/Tweet"
      }
    }
  },
  "required" : [ "person", "tweets" ],
  "additionalProperties" : false,
  "$defs" : {
    "Person" : {
      "type" : "object",
      "properties" : {
        "age" : {
          "type" : "integer"
        },
        "friends" : {
          "type" : "array",
          "items" : {
            "$ref" : "#/$defs/Person"
          }
        },
        "name" : {
          "type" : "string"
        },
        "tweets" : {
          "type" : "array",
          "items" : {
            "$ref" : "#/$defs/Tweet"
          }
        }
      },
      "additionalProperties" : false
    },
    "Tweet" : {
      "type" : "object",
      "properties" : {
        "text" : {
          "type" : "string"
        }
      },
      "required" : [ "text" ],
      "additionalProperties" : false
    }
  }
}`,
		},
		{
			name: "Test Person",
			in:   Person{},
			want: `{
  "type": "object",
  "properties": {
    "age": {
      "type": "integer"
    },
    "friends": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/Person"
      }
    },
    "name": {
      "type": "string"
    },
    "tweets": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/Tweet"
      }
    }
  },
  "additionalProperties": false,
  "$defs": {
    "Person": {
      "type": "object",
      "properties": {
        "age": {
          "type": "integer"
        },
        "friends": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/Person"
          }
        },
        "name": {
          "type": "string"
        },
        "tweets": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/Tweet"
          }
        }
      },
      "additionalProperties": false
    },
    "Tweet": {
      "type": "object",
      "properties": {
        "text": {
          "type": "string"
        }
      },
      "required": [
        "text"
      ],
      "additionalProperties": false
    }
  }
}`,
		},
		{
			name: "Test MyStructuredResponse",
			in:   MyStructuredResponse{},
			want: `{
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
    "kebab_case",
    "snake_case"
  ],
  "additionalProperties": false
}`,
		},
		{
			name: "Test User",
			in:   User{},
			want: `{
  "type": "object",
  "properties": {
    "id": {
      "type": "integer"
    },
    "name": {
      "type": "string"
    },
    "orders": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/Order"
      }
    }
  },
  "additionalProperties": false,
  "$defs": {
    "Order": {
      "type": "object",
      "properties": {
        "amount": {
          "type": "number"
        },
        "buyer": {
          "$ref": "#/$defs/User"
        },
        "id": {
          "type": "integer"
        }
      },
      "additionalProperties": false
    },
    "User": {
      "type": "object",
      "properties": {
        "id": {
          "type": "integer"
        },
        "name": {
          "type": "string"
        },
        "orders": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/Order"
          }
        }
      },
      "additionalProperties": false
    }
  }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantBytes := []byte(tt.want)

			schema, err := jsonschema.GenerateSchemaForType(tt.in)
			if err != nil {
				t.Errorf("Failed to generate schema: error = %v", err)
				return
			}

			var want map[string]interface{}
			err = json.Unmarshal(wantBytes, &want)
			if err != nil {
				t.Errorf("Failed to Unmarshal JSON: error = %v", err)
				return
			}

			got := structToMap(t, schema)
			gotPtr := structToMap(t, &schema)

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
