package openai_test

import (
	"encoding/json"
	"reflect"
	"testing"

	openai "github.com/meguminnnnnnnnn/go-openai/internal"
	"github.com/meguminnnnnnnnn/go-openai/internal/test/checks"
	"github.com/stretchr/testify/assert"
)

func TestJSONUnmarshaler_Normal(t *testing.T) {
	jm := &openai.JSONUnmarshaler{}
	data := []byte(`{"key":"value"}`)
	var v map[string]string

	err := jm.Unmarshal(data, &v)
	checks.NoError(t, err)
	if v["key"] != "value" {
		t.Fatal("unmarshal result mismatch")
	}
}

func TestJSONUnmarshaler_InvalidJSON(t *testing.T) {
	jm := &openai.JSONUnmarshaler{}
	data := []byte(`{invalid}`)
	var v map[string]interface{}

	err := jm.Unmarshal(data, &v)
	checks.HasError(t, err, "should return error for invalid JSON")
}

func TestJSONUnmarshaler_EmptyInput(t *testing.T) {
	jm := &openai.JSONUnmarshaler{}
	var v interface{}

	err := jm.Unmarshal(nil, &v)
	checks.HasError(t, err, "should return error for nil input")
}

func TestUnmarshalExtraFields(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1"`
		Field2 int
		Field3 struct {
			Field4 string `json:"field4"`
		} `json:"field3"`
	}

	testData := []byte(`{"field1":"value1","Field2":2,"field3":{"field4":"value4"},"extraField1":"extraValue1"}`)
	testStruct := &TestStruct{}
	extra, err := openai.UnmarshalExtraFields(reflect.TypeOf(testStruct), testData)
	assert.NoError(t, err)
	assert.Len(t, extra, 1)

	var extraValue1 string
	err = json.Unmarshal(extra["extraField1"], &extraValue1)
	assert.NoError(t, err)

	assert.Equal(t, "extraValue1", extraValue1)
}
