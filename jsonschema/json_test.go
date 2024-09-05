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
		def  any
		want string
	}{
		{
			name: "Test with empty Definition",
			def:  struct{}{},
			want: `{
	"type":"object",
	"additionalProperties":false
}`,
		},
		{
			name: "Test with nested Definition properties",
			def: struct {
				User struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				} `json:"user"`
			}{},
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
   },
   "additionalProperties":false,
   "required":["user"]
}`,
		},
		{
			name: "Test with complex nested Definition",
			def: struct {
				User struct {
					Name    string `json:"name"`
					Age     int    `json:"age"`
					Address struct {
						City    string `json:"city"`
						Country string `json:"country"`
					} `json:"address"`
				} `json:"user"`
			}{},
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
         },
		 "additionalProperties":false,		
		 "required":["name","age","address"]
      }
   },
   "additionalProperties":false,
   "required":["user"]
}`,
		},
		{
			name: "Test with Array type Definition",
			def:  []string{},
			want: `{
   "type":"array",
   "items":{
      "type":"string"
   }
}`,
		},
		{
			name: "Test order prevention",
			def: struct {
				C string `json:"c"`
				A int    `json:"a"`
				B bool   `json:"b"`
			}{},
			want: `{
   "type":"object",
   "properties":{
      "c":{
         "type":"string"
      },
      "a":{
         "type":"integer"
      },
      "b":{
         "type":"boolean"
      }
   },
   "additionalProperties":false,
   "required":["c","a","b"]
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantBytes := []byte(tt.want)
			var wantDef jsonschema.Definition
			err := json.Unmarshal(wantBytes, &wantDef)
			if err != nil {
				t.Errorf("Failed to Unmarshal JSON: error = %v", err)
				return
			}
			schema, err := jsonschema.GenerateSchemaForType(tt.def)
			if err != nil {
				t.Errorf("Failed to Unmarshal JSON: error = %v", err)
				return
			}
			gotBytes, err := schema.MarshalJSON()
			if err != nil {
				t.Errorf("Failed to Marshal JSON: error = %v", err)
				return
			}
			var gotDef jsonschema.Definition
			err = json.Unmarshal(gotBytes, &gotDef)
			if err != nil {
				t.Errorf("Failed to Unmarshal JSON: error = %v", err)
				return
			}
			if !reflect.DeepEqual(wantDef, gotDef) {
				t.Errorf("Definition.MarshalJSON() = %v, want %v", gotDef, wantDef)
			}
		})
	}
}
