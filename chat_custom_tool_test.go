package openai_test

import (
	"encoding/json"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestToolCallCustomTypeDeserialization(t *testing.T) {
	var resp openai.ChatCompletionResponse
	err := json.Unmarshal([]byte(`{
		"choices": [{
			"message": {
				"role": "assistant",
				"tool_calls": [{
					"id": "call_abc123",
					"type": "custom",
					"custom": {"name": "my_custom_tool", "input": "some input data"}
				}]
			},
			"finish_reason": "tool_calls"
		}]
	}`), &resp)
	checks.NoError(t, err, "unmarshal chat completion response")
	if len(resp.Choices) != 1 || len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("expected one tool call, got %#v", resp.Choices)
	}

	call := resp.Choices[0].Message.ToolCalls[0]
	if call.Type != openai.ToolTypeCustom {
		t.Fatalf("expected custom tool call type, got %q", call.Type)
	}
	if call.Custom == nil {
		t.Fatal("expected custom tool call payload to be decoded")
	}
	if call.Custom.Name != "my_custom_tool" || call.Custom.Input != "some input data" {
		t.Fatalf("unexpected custom tool call payload: %#v", call.Custom)
	}

	serialized, err := json.Marshal(call)
	checks.NoError(t, err, "marshal custom tool call")
	var roundTrip map[string]any
	checks.NoError(t, json.Unmarshal(serialized, &roundTrip), "unmarshal serialized custom tool call")
	custom, ok := roundTrip["custom"].(map[string]any)
	if !ok {
		t.Fatalf("expected custom key after round trip, got %s", serialized)
	}
	if custom["name"] != "my_custom_tool" || custom["input"] != "some input data" {
		t.Fatalf("unexpected custom key after round trip: %#v", custom)
	}
}

func TestToolDefinitionCustomType(t *testing.T) {
	tool := openai.Tool{
		Type: openai.ToolTypeCustom,
		Custom: &openai.CustomToolDefinition{
			Name:        "my_custom_tool",
			Description: "Formats output",
			Format: &openai.CustomToolFormat{
				Type: openai.CustomToolFormatTypeGrammar,
				Grammar: &openai.CustomToolGrammar{
					Definition: "start: line+",
					Syntax:     "lark",
				},
			},
		},
	}

	serialized, err := json.Marshal(tool)
	checks.NoError(t, err, "marshal custom tool definition")
	var got map[string]any
	checks.NoError(t, json.Unmarshal(serialized, &got), "unmarshal custom tool definition")
	custom, ok := got["custom"].(map[string]any)
	if !ok {
		t.Fatalf("expected custom key in tool definition, got %s", serialized)
	}
	if custom["name"] != "my_custom_tool" || custom["description"] != "Formats output" {
		t.Fatalf("unexpected custom definition payload: %#v", custom)
	}
	format, ok := custom["format"].(map[string]any)
	if !ok || format["type"] != "grammar" {
		t.Fatalf("unexpected custom format payload: %#v", custom)
	}
	grammar, ok := format["grammar"].(map[string]any)
	if !ok || grammar["definition"] != "start: line+" || grammar["syntax"] != "lark" {
		t.Fatalf("unexpected custom grammar payload: %#v", custom)
	}

	var decoded openai.Tool
	checks.NoError(t, json.Unmarshal(serialized, &decoded), "decode custom tool definition")
	if decoded.Custom == nil || decoded.Custom.Name != "my_custom_tool" {
		t.Fatalf("expected custom definition to survive round trip, got %#v", decoded.Custom)
	}
}

func TestToolParametersCannotOverrideCustom(t *testing.T) {
	_, err := json.Marshal(openai.Tool{
		Type:   openai.ToolTypeCustom,
		Custom: &openai.CustomToolDefinition{Name: "my_custom_tool"},
		Parameters: map[string]any{
			"custom": map[string]any{"name": "other"},
		},
	})
	checks.HasError(t, err)
}
