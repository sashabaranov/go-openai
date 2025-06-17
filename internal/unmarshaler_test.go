package openai_test

import (
	"testing"

	openai "github.com/sashabaranov/go-openai/internal"
	"github.com/sashabaranov/go-openai/internal/test/checks"
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
