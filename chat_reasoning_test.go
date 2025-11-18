package openai_test

import (
	"encoding/json"
	"testing"

	"github.com/sashabaranov/go-openai"
)

// TestChatCompletionStreamChoiceDelta_ReasoningFieldSupport tests that both
// reasoning_content and reasoning fields are properly supported in streaming responses.
func TestChatCompletionStreamChoiceDelta_ReasoningFieldSupport(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected string
	}{
		{
			name:     "DeepSeek style - reasoning_content",
			jsonData: `{"role":"assistant","content":"Hello","reasoning_content":"This is my reasoning"}`,
			expected: "This is my reasoning",
		},
		{
			name:     "OpenAI style - reasoning",
			jsonData: `{"role":"assistant","content":"Hello","reasoning":"This is my reasoning"}`,
			expected: "This is my reasoning",
		},
		{
			name:     "Both fields present - reasoning_content takes priority",
			jsonData: `{"role":"assistant","content":"Hello","reasoning_content":"Priority reasoning","reasoning":"Fallback reasoning"}`,
			expected: "Priority reasoning",
		},
		{
			name:     "Only reasoning field",
			jsonData: `{"role":"assistant","reasoning":"Only reasoning field"}`,
			expected: "Only reasoning field",
		},
		{
			name:     "No reasoning fields",
			jsonData: `{"role":"assistant","content":"Hello"}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var delta openai.ChatCompletionStreamChoiceDelta
			err := json.Unmarshal([]byte(tt.jsonData), &delta)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if delta.ReasoningContent != tt.expected {
				t.Errorf("Expected ReasoningContent to be %q, got %q", tt.expected, delta.ReasoningContent)
			}
		})
	}
}

// TestChatCompletionMessage_ReasoningFieldSupport tests that both
// reasoning_content and reasoning fields are properly supported in chat completion messages.
func TestChatCompletionMessage_ReasoningFieldSupport(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected string
	}{
		{
			name:     "DeepSeek style - reasoning_content",
			jsonData: `{"role":"assistant","content":"Hello","reasoning_content":"This is my reasoning"}`,
			expected: "This is my reasoning",
		},
		{
			name:     "OpenAI style - reasoning",
			jsonData: `{"role":"assistant","content":"Hello","reasoning":"This is my reasoning"}`,
			expected: "This is my reasoning",
		},
		{
			name:     "Both fields present - reasoning_content takes priority",
			jsonData: `{"role":"assistant","content":"Hello","reasoning_content":"Priority reasoning","reasoning":"Fallback reasoning"}`,
			expected: "Priority reasoning",
		},
		{
			name:     "Only reasoning field",
			jsonData: `{"role":"assistant","reasoning":"Only reasoning field"}`,
			expected: "Only reasoning field",
		},
		{
			name:     "No reasoning fields",
			jsonData: `{"role":"assistant","content":"Hello"}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg openai.ChatCompletionMessage
			err := json.Unmarshal([]byte(tt.jsonData), &msg)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if msg.ReasoningContent != tt.expected {
				t.Errorf("Expected ReasoningContent to be %q, got %q", tt.expected, msg.ReasoningContent)
			}
		})
	}
}

// TestChatCompletionMessage_MultiContent_ReasoningFieldSupport tests reasoning field support
// with MultiContent messages.
func TestChatCompletionMessage_MultiContent_ReasoningFieldSupport(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected string
	}{
		{
			name:     "MultiContent with reasoning_content",
			jsonData: `{"role":"assistant","content":[{"type":"text","text":"Hello"}],"reasoning_content":"Multi reasoning"}`,
			expected: "Multi reasoning",
		},
		{
			name:     "MultiContent with reasoning",
			jsonData: `{"role":"assistant","content":[{"type":"text","text":"Hello"}],"reasoning":"Multi reasoning"}`,
			expected: "Multi reasoning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg openai.ChatCompletionMessage
			err := json.Unmarshal([]byte(tt.jsonData), &msg)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if msg.ReasoningContent != tt.expected {
				t.Errorf("Expected ReasoningContent to be %q, got %q", tt.expected, msg.ReasoningContent)
			}
		})
	}
}

// TestChatCompletionStreamChoiceDelta_MarshalJSON tests that marshaling preserves
// the reasoning_content field name.
func TestChatCompletionStreamChoiceDelta_MarshalJSON(t *testing.T) {
	delta := openai.ChatCompletionStreamChoiceDelta{
		Role:             "assistant",
		Content:          "Hello",
		ReasoningContent: "Test reasoning",
	}

	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("Failed to marshal delta: %v", err)
	}

	// Verify that it's marshaled as reasoning_content
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if _, hasReasoning := result["reasoning"]; hasReasoning {
		t.Error("Marshaled JSON should not contain 'reasoning' field")
	}

	if reasoningContent, ok := result["reasoning_content"].(string); !ok || reasoningContent != "Test reasoning" {
		t.Errorf("Expected reasoning_content to be 'Test reasoning', got %v", result["reasoning_content"])
	}
}

// TestRealWorldStreamingResponse tests parsing a real-world streaming response
// with reasoning field (similar to the new_api_output.txt format).
func TestRealWorldStreamingResponse(t *testing.T) {
	// Simulate a chunk from the new_api_output.txt file
	jsonData := `{
		"id":"gen-1763431956-test",
		"provider":"Azure",
		"model":"openai/gpt-5",
		"object":"chat.completion.chunk",
		"created":1763431956,
		"choices":[{
			"index":0,
			"delta":{
				"role":"assistant",
				"content":"",
				"reasoning":"quantum"
			},
			"finish_reason":null,
			"logprobs":null
		}]
	}`

	var response openai.ChatCompletionStreamResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal streaming response: %v", err)
	}

	if len(response.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(response.Choices))
	}

	delta := response.Choices[0].Delta
	if delta.ReasoningContent != "quantum" {
		t.Errorf("Expected ReasoningContent to be 'quantum', got %q", delta.ReasoningContent)
	}

	if delta.Role != "assistant" {
		t.Errorf("Expected Role to be 'assistant', got %q", delta.Role)
	}
}
