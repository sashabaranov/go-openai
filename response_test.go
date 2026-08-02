package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestCreateResponse(t *testing.T) { //nolint:gocognit
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/responses", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var request map[string]any
		checks.NoError(t, json.NewDecoder(r.Body).Decode(&request), "decode request")
		if request["model"] != "extra-model" {
			t.Errorf("expected ExtraBody to override model, got %v", request["model"])
		}
		if request["input"] != "Hello" {
			t.Errorf("expected input Hello, got %v", request["input"])
		}
		if request["store"] != false {
			t.Errorf("expected store=false, got %v", request["store"])
		}
		if request["parallel_tool_calls"] != false {
			t.Errorf("expected parallel_tool_calls=false, got %v", request["parallel_tool_calls"])
		}
		if request["custom_option"] != "custom-value" {
			t.Errorf("expected custom ExtraBody value, got %v", request["custom_option"])
		}

		tools, ok := request["tools"].([]any)
		if !ok || len(tools) != 1 {
			t.Fatalf("expected one tool, got %#v", request["tools"])
		}
		tool, ok := tools[0].(map[string]any)
		if !ok {
			t.Fatalf("expected object tool, got %#v", tools[0])
		}
		if tool["type"] != "web_search" || tool["search_context_size"] != "low" {
			t.Errorf("unexpected tool payload: %#v", tool)
		}

		w.Header().Set(xCustomHeader, xCustomHeaderValue)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{
			"id":"resp_123",
			"object":"response",
			"created_at":1741476542,
			"status":"completed",
			"model":"gpt-4o",
			"output":[{
				"id":"msg_123",
				"type":"message",
				"status":"completed",
				"role":"assistant",
				"content":[{"type":"output_text","text":"Hello back","annotations":[]}]
			}],
			"usage":{
				"input_tokens":7,
				"input_tokens_details":{"cached_tokens":2,"cache_write_tokens":1},
				"output_tokens":3,
				"output_tokens_details":{"reasoning_tokens":1},
				"total_tokens":10
			}
		}`))
		checks.NoError(t, err, "write response")
	})

	store := false
	parallelToolCalls := false
	response, err := client.CreateResponse(context.Background(), openai.CreateResponseRequest{
		Model:             openai.GPT4o,
		Input:             "Hello",
		Store:             &store,
		ParallelToolCalls: &parallelToolCalls,
		Tools: []openai.Tool{{
			Type: openai.ToolTypeWebSearch,
			Parameters: map[string]any{
				"search_context_size": "low",
			},
		}},
		ExtraBody: map[string]any{
			"model":         "extra-model",
			"custom_option": "custom-value",
		},
	})
	checks.NoError(t, err, "CreateResponse returned error")

	if response.ID != "resp_123" || response.Created != 1741476542 {
		t.Errorf("unexpected response identity: %#v", response)
	}
	if response.Status != openai.ResponseStatusCompleted {
		t.Errorf("expected completed status, got %q", response.Status)
	}
	if response.Usage == nil || response.Usage.TotalTokens != 10 {
		t.Errorf("unexpected usage: %#v", response.Usage)
	}
	if response.GetOutputText() != "Hello back" {
		t.Errorf("expected aggregated output text, got %q", response.GetOutputText())
	}
	if response.Header().Get(xCustomHeader) != xCustomHeaderValue {
		t.Errorf("expected response header to be retained")
	}
}

func TestCreateResponseUsesOutputTextConvenienceField(t *testing.T) {
	response := openai.CreateResponseResponse{
		OutputText: "convenience text",
		Output: []any{map[string]any{
			"type": "message",
			"content": []any{
				map[string]any{"type": "output_text", "text": "nested text"},
			},
		}},
	}
	if response.GetOutputText() != "convenience text" {
		t.Errorf("expected output_text field to take precedence")
	}
}

func TestCreateResponseRejectsStreaming(t *testing.T) {
	client := openai.NewClient("test")
	_, err := client.CreateResponse(context.Background(), openai.CreateResponseRequest{Stream: true})
	checks.ErrorIs(t, err, openai.ErrResponseStreamNotSupported, "expected stream guard")
}

func TestCreateResponseRequestMarshalError(t *testing.T) {
	client := openai.NewClient("test")
	_, err := client.CreateResponse(context.Background(), openai.CreateResponseRequest{
		Input: make(chan int),
	})
	checks.HasError(t, err, "expected JSON marshal error")
}

func TestCreateResponseError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/responses", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, err := w.Write([]byte(`{"error":{"message":"slow down","type":"rate_limit_error","code":"rate_limit"}}`))
		checks.NoError(t, err, "write error response")
	})

	_, err := client.CreateResponse(context.Background(), openai.CreateResponseRequest{
		Model: openai.GPT4o,
		Input: "Hello",
	})
	var apiError *openai.APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiError.HTTPStatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", apiError.HTTPStatusCode)
	}
}

func TestResponseFunctionAndBuiltinToolMarshalJSON(t *testing.T) {
	functionTool := openai.NewResponseFunctionTool(openai.FunctionDefinition{
		Name:        "get_weather",
		Description: "Get the weather",
		Strict:      true,
		Parameters: map[string]any{
			"type": "object",
		},
	})
	encoded, err := json.Marshal(functionTool)
	checks.NoError(t, err, "marshal response function tool")

	var got map[string]any
	checks.NoError(t, json.Unmarshal(encoded, &got), "unmarshal response function tool")
	if got["type"] != "function" || got["name"] != "get_weather" || got["strict"] != true {
		t.Errorf("unexpected function tool: %#v", got)
	}
	if _, exists := got["function"]; exists {
		t.Errorf("Responses API function tool must use inline fields: %#v", got)
	}

	chatTool := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:       "get_weather",
			Parameters: map[string]any{"type": "object"},
		},
	}
	encoded, err = json.Marshal(chatTool)
	checks.NoError(t, err, "marshal chat function tool")
	checks.NoError(t, json.Unmarshal(encoded, &got), "unmarshal chat function tool")
	if _, exists := got["function"]; !exists {
		t.Errorf("Chat Completions tool lost nested function: %#v", got)
	}

	_, err = json.Marshal(openai.Tool{
		Type:       openai.ToolTypeWebSearch,
		Parameters: map[string]any{"type": "function"},
	})
	checks.HasError(t, err, "expected reserved tool property error")
}

func TestResponseLifecycleEndpoints(t *testing.T) { //nolint:gocognit
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/responses/resp_123", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			values := r.URL.Query()
			expectedInclude := []string{
				string(openai.ResponseIncludeFileSearchCallResults),
				string(openai.ResponseIncludeReasoningEncryptedContent),
			}
			if !reflect.DeepEqual(values["include"], expectedInclude) {
				t.Errorf("unexpected include query: %#v", values["include"])
			}
			if values.Get("include_obfuscation") != "false" || values.Get("starting_after") != "2" {
				t.Errorf("unexpected retrieve query: %s", r.URL.RawQuery)
			}
			_, err := w.Write([]byte(`{"id":"resp_123","object":"response","model":"gpt-4o","output":[]}`))
			checks.NoError(t, err, "write retrieve response")
		case http.MethodDelete:
			_, err := w.Write([]byte(`{"id":"resp_123","object":"response","deleted":true}`))
			checks.NoError(t, err, "write delete response")
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})
	server.RegisterHandler("/v1/responses/resp_123/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		_, err := w.Write([]byte(`{"id":"resp_123","object":"response","status":"cancelled","output":[]}`))
		checks.NoError(t, err, "write cancel response")
	})
	server.RegisterHandler("/v1/responses/resp_123/input_items", func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		if values.Get("after") != "item_1" || values.Get("limit") != "10" || values.Get("order") != "desc" {
			t.Errorf("unexpected list query: %s", r.URL.RawQuery)
		}
		_, err := w.Write([]byte(`{
			"object":"list","data":[{"type":"message","role":"user"}],
			"first_id":"item_2","last_id":"item_2","has_more":false
		}`))
		checks.NoError(t, err, "write input items response")
	})
	server.RegisterHandler("/v1/responses/input_tokens", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		_, err := w.Write([]byte(`{"object":"response.input_tokens","input_tokens":42}`))
		checks.NoError(t, err, "write token count response")
	})
	server.RegisterHandler("/v1/responses/compact", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		_, err := w.Write([]byte(`{
			"id":"cmp_123","object":"response.compaction","created_at":1741476542,
			"output":[{"type":"compaction","encrypted_content":"encrypted"}]
		}`))
		checks.NoError(t, err, "write compact response")
	})

	includeObfuscation := false
	startingAfter := 2
	retrieved, err := client.GetResponse(context.Background(), "resp_123", openai.RetrieveResponseOptions{
		Include: []openai.ResponseInclude{
			openai.ResponseIncludeFileSearchCallResults,
			openai.ResponseIncludeReasoningEncryptedContent,
		},
		IncludeObfuscation: &includeObfuscation,
		StartingAfter:      &startingAfter,
	})
	checks.NoError(t, err, "GetResponse returned error")
	if retrieved.ID != "resp_123" {
		t.Errorf("unexpected retrieved response: %#v", retrieved)
	}

	deleted, err := client.DeleteResponse(context.Background(), "resp_123")
	checks.NoError(t, err, "DeleteResponse returned error")
	if !deleted.Deleted {
		t.Errorf("expected response to be deleted")
	}

	cancelled, err := client.CancelResponse(context.Background(), "resp_123")
	checks.NoError(t, err, "CancelResponse returned error")
	if cancelled.Status != openai.ResponseStatusCancelled {
		t.Errorf("unexpected cancelled status: %q", cancelled.Status)
	}

	items, err := client.ListResponseInputItems(
		context.Background(),
		"resp_123",
		openai.ResponseInputItemsListOptions{After: "item_1", Limit: 10, Order: "desc"},
	)
	checks.NoError(t, err, "ListResponseInputItems returned error")
	if len(items.Data) != 1 || items.FirstID != "item_2" {
		t.Errorf("unexpected input items response: %#v", items)
	}

	tokens, err := client.CountResponseInputTokens(context.Background(), openai.CreateResponseRequest{
		Model: openai.GPT4o,
		Input: "Hello",
	})
	checks.NoError(t, err, "CountResponseInputTokens returned error")
	if tokens.InputTokens != 42 {
		t.Errorf("expected 42 input tokens, got %d", tokens.InputTokens)
	}

	compaction, err := client.CompactResponse(context.Background(), openai.CreateResponseRequest{
		Model: openai.GPT4o,
		Input: "Hello",
	})
	checks.NoError(t, err, "CompactResponse returned error")
	if compaction.ID != "cmp_123" || len(compaction.Output) != 1 {
		t.Errorf("unexpected compaction response: %#v", compaction)
	}
}
