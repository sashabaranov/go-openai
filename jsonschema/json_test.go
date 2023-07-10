package jsonschema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	. "github.com/sashabaranov/go-openai/jsonschema"
)

func TestDefinition_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		def  Definition
		want string
	}{
		{
			name: "Test with empty Definition",
			def:  Definition{},
			want: `{}`,
		},
		{
			name: "Test with Definition properties set",
			def: Definition{
				Type:        String,
				Description: "A string type",
				Properties: map[string]Definition{
					"name": {
						Type: String,
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
			def: Definition{
				Type: Object,
				Properties: map[string]Definition{
					"user": {
						Type: Object,
						Properties: map[string]Definition{
							"name": {
								Type: String,
							},
							"age": {
								Type: Integer,
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
			def: Definition{
				Type: Object,
				Properties: map[string]Definition{
					"user": {
						Type: Object,
						Properties: map[string]Definition{
							"name": {
								Type: String,
							},
							"age": {
								Type: Integer,
							},
							"address": {
								Type: Object,
								Properties: map[string]Definition{
									"city": {
										Type: String,
									},
									"country": {
										Type: String,
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
			def: Definition{
				Type: Array,
				Items: &Definition{
					Type: String,
				},
				Properties: map[string]Definition{
					"name": {
						Type: String,
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
			gotBytes, err := json.Marshal(&tt.def)
			if err != nil {
				t.Errorf("Failed to Marshal JSON: error = %v", err)
				return
			}

			var got map[string]interface{}
			err = json.Unmarshal(gotBytes, &got)
			if err != nil {
				t.Errorf("Failed to Unmarshal JSON: error =  %v", err)
				return
			}

			wantBytes := []byte(tt.want)
			var want map[string]interface{}
			err = json.Unmarshal(wantBytes, &want)
			if err != nil {
				t.Errorf("Failed to Unmarshal JSON: error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, want)
			}
		})
	}
}
