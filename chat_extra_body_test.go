package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

// These tests cover the fork-only feature: ChatCompletionRequest.ExtraBody is
// merged into the top-level JSON body of chat completion requests (the same
// semantics as EmbeddingRequest.ExtraBody upstream), so vendor-specific
// parameters (qwen, deepseek, doubao, ...) can be sent without changing the
// request struct. Keys in ExtraBody override the struct fields.

const (
	chatCompletionsPath = "/v1/chat/completions"
	wantTopK            = 5
	wantThinkingBudget  = 1024
	wantMaxTokens       = 7
)

// chatBodyRecorder captures the decoded JSON body of the last chat completion
// request received by the test server.
type chatBodyRecorder struct {
	body map[string]any
	err  error
}

func (rec *chatBodyRecorder) record(r *http.Request) {
	rec.body = nil
	rec.err = json.NewDecoder(r.Body).Decode(&rec.body)
}

func (rec *chatBodyRecorder) assertNoNestedExtraBody(t *testing.T) {
	checks.NoError(t, rec.err, "decode request body")
	if _, exists := rec.body["extra_body"]; exists {
		t.Fatalf("extra_body must be merged into the top level, got nested field: %v", rec.body["extra_body"])
	}
}

func (rec *chatBodyRecorder) syncHandler(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rec.record(r)
		resp := openai.ChatCompletionResponse{ID: "1", Object: "chat.completion", Model: openai.GPT4oMini}
		checks.NoError(t, json.NewEncoder(w).Encode(resp), "encode response")
	}
}

func TestChatCompletionExtraBodyMergedIntoRequest(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	rec := &chatBodyRecorder{}
	server.RegisterHandler(chatCompletionsPath, rec.syncHandler(t))

	// Consumers override `messages` this way to inject message attributes the
	// struct cannot express (e.g. DashScope cache_control markers).
	overrideMessages := []any{
		map[string]any{
			"role": "user",
			"content": []any{
				map[string]any{
					"type":          "text",
					"text":          "Hello!",
					"cache_control": map[string]any{"type": "ephemeral"},
				},
			},
		},
	}

	_, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "Hello!"}},
		ExtraBody: map[string]any{
			"top_k":    wantTopK,
			"thinking": map[string]any{"type": "enabled"},
			"messages": overrideMessages,
		},
	})
	checks.NoError(t, err, "CreateChatCompletion error")
	rec.assertNoNestedExtraBody(t)

	if rec.body["model"] != openai.GPT4oMini {
		t.Errorf("struct fields must survive the merge, model = %v", rec.body["model"])
	}
	if rec.body["stream"] != false {
		t.Errorf("non-streaming request must carry stream=false, got %v", rec.body["stream"])
	}
	// JSON numbers decode into float64.
	if rec.body["top_k"] != float64(wantTopK) {
		t.Errorf("top_k = %v, want %d", rec.body["top_k"], wantTopK)
	}
	thinking, ok := rec.body["thinking"].(map[string]any)
	if !ok || thinking["type"] != "enabled" {
		t.Errorf("thinking = %v, want {type: enabled}", rec.body["thinking"])
	}
	if !reflect.DeepEqual(rec.body["messages"], overrideMessages) {
		t.Errorf("ExtraBody keys must override struct fields, messages = %v", rec.body["messages"])
	}
}

func TestChatCompletionStreamExtraBodyMergedIntoRequest(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	rec := &chatBodyRecorder{}
	server.RegisterHandler(chatCompletionsPath, func(w http.ResponseWriter, r *http.Request) {
		rec.record(r)
		w.Header().Set("Content-Type", "text/event-stream")
		chunk := `{"id":"1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"ok"}}]}`
		_, err := w.Write([]byte("data: " + chunk + "\n\ndata: [DONE]\n\n"))
		checks.NoError(t, err, "write stream")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "Hello!"}},
		ExtraBody: map[string]any{
			"enable_thinking": true,
			"thinking_budget": wantThinkingBudget,
		},
	})
	checks.NoError(t, err, "CreateChatCompletionStream error")
	defer stream.Close()

	_, err = stream.Recv()
	checks.NoError(t, err, "stream.Recv error")
	rec.assertNoNestedExtraBody(t)

	if rec.body["stream"] != true {
		t.Errorf("streaming request must carry stream=true, got %v", rec.body["stream"])
	}
	if rec.body["enable_thinking"] != true {
		t.Errorf("enable_thinking = %v, want true", rec.body["enable_thinking"])
	}
	if rec.body["thinking_budget"] != float64(wantThinkingBudget) {
		t.Errorf("thinking_budget = %v, want %d", rec.body["thinking_budget"], wantThinkingBudget)
	}
}

func TestChatCompletionWithoutExtraBodyIsUnchanged(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	rec := &chatBodyRecorder{}
	server.RegisterHandler(chatCompletionsPath, rec.syncHandler(t))

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT4oMini,
		Messages:  []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "Hello!"}},
		MaxTokens: wantMaxTokens,
	}
	_, err := client.CreateChatCompletion(context.Background(), req)
	checks.NoError(t, err, "CreateChatCompletion error")
	rec.assertNoNestedExtraBody(t)

	messages, ok := rec.body["messages"].([]any)
	if !ok || len(messages) != len(req.Messages) {
		t.Errorf("messages = %v, want the %d struct message(s)", rec.body["messages"], len(req.Messages))
	}
	if rec.body["max_tokens"] != float64(wantMaxTokens) {
		t.Errorf("max_tokens = %v, want %d", rec.body["max_tokens"], wantMaxTokens)
	}
}
