package openai_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestResponseMultiAgentRequestAndStream(t *testing.T) { //nolint:gocognit
	for _, streaming := range []bool{false, true} {
		t.Run(map[bool]string{false: "response", true: "stream"}[streaming], func(t *testing.T) {
			client, server, teardown := setupOpenAITestServer()
			defer teardown()
			server.RegisterHandler("/v1/responses", func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("OpenAI-Beta") != "responses_multi_agent=v1" {
					t.Errorf("missing beta header: %v", r.Header)
				}
				var body map[string]json.RawMessage
				checks.NoError(t, json.NewDecoder(r.Body).Decode(&body), "decode request")
				if _, exists := body["Betas"]; exists {
					t.Error("Betas must not be in the body")
				}
				if _, exists := body["betas"]; exists {
					t.Error("betas must not be in the body")
				}
				var config openai.ResponseMultiAgent
				checks.NoError(t, json.Unmarshal(body["multi_agent"], &config), "decode config")
				if !config.Enabled || config.MaxConcurrentSubagents != 2 {
					t.Errorf("unexpected multi_agent: %+v", config)
				}
				if streaming {
					w.Header().Set("Content-Type", "text/event-stream")
					_, err := io.WriteString(w, "data: "+`{"type":"response.output_item.done",`+
						`"agent":{"agent_name":"/root/research"},`+
						`"item":{"type":"function_call","call_id":"call_1","name":"lookup","async":true,`+
						`"agent":{"agent_name":"/root/research"}}}`+"\n\n")
					checks.NoError(t, err, "write SSE")
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, err := io.WriteString(w, `{"id":"resp_multi","output":[
					{"type":"message","agent":{"agent_name":"/root/worker"},"phase":"final_answer",
					"content":[{"type":"output_text","text":"worker output"}]},
					{"type":"message","agent":{"agent_name":"/root"},"phase":"final_answer",
					"content":[{"type":"output_text","text":"Final answer"}]}]}`)
				checks.NoError(t, err, "write response")
			})
			request := openai.CreateResponseRequest{
				Model: openai.GPT6Dot1Sol, Input: "Compare two designs.",
				Betas:      []string{openai.ResponseBetaMultiAgent},
				MultiAgent: &openai.ResponseMultiAgent{Enabled: true, MaxConcurrentSubagents: 2},
			}
			if !streaming {
				response, err := client.CreateResponse(context.Background(), request)
				checks.NoError(t, err, "create response")
				if response.GetFinalOutputText() != "Final answer" {
					t.Errorf("unexpected final text: %q", response.GetFinalOutputText())
				}
				return
			}
			stream, err := client.CreateResponseStream(context.Background(), request)
			checks.NoError(t, err, "create stream")
			defer stream.Close()
			event, err := stream.Recv()
			checks.NoError(t, err, "receive event")
			if event.Agent == nil || event.Agent.AgentName != "/root/research" ||
				event.Item == nil || !event.Item.Async || event.Item.CallID != "call_1" {
				t.Fatalf("lost agent or async metadata: %+v", event)
			}
			if !strings.Contains(string(event.Raw), "lookup") {
				t.Error("raw event not retained")
			}
		})
	}
}

func TestResponseConfigurationAndAsyncToolPayload(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/responses", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("OpenAI-Beta") != "" {
			t.Error("ordinary requests must not opt in to a beta")
		}
		var body struct {
			Model       string                               `json:"model"`
			ServiceTier string                               `json:"service_tier"`
			Reasoning   openai.ResponseReasoning             `json:"reasoning"`
			Input       []openai.ResponseConfigurationUpdate `json:"input"`
			Tools       []struct {
				Type     string `json:"type"`
				Name     string `json:"name"`
				Async    bool   `json:"async"`
				Function any    `json:"function"`
			} `json:"tools"`
		}
		checks.NoError(t, json.NewDecoder(r.Body).Decode(&body), "decode request")
		if body.Model != "gpt-6-astra" || body.ServiceTier != "ultrafast" || body.Reasoning.Effort != "low" {
			t.Errorf("unexpected request configuration: %+v", body)
		}
		if len(body.Input) != 1 || body.Input[0].Type != "configuration_update" || body.Input[0].Reasoning.Effort != "high" {
			t.Errorf("unexpected configuration update: %+v", body.Input)
		}
		if len(body.Tools) != 1 || body.Tools[0].Type != "function" || !body.Tools[0].Async ||
			body.Tools[0].Name != "lookup" || body.Tools[0].Function != nil {
			t.Errorf("unexpected async tool: %+v", body.Tools)
		}
		w.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(w, `{"id":"resp_config","output":[]}`)
		checks.NoError(t, err, "write response")
	})
	_, err := client.CreateResponse(context.Background(), openai.CreateResponseRequest{
		Model: openai.GPT6Astra, ServiceTier: string(openai.ServiceTierUltrafast),
		Reasoning: &openai.ResponseReasoning{Effort: openai.ReasoningEffortLow},
		Input: []any{openai.ResponseConfigurationUpdate{
			Type: "configuration_update", Reasoning: openai.ResponseReasoning{Effort: openai.ReasoningEffortHigh},
		}},
		Tools: []openai.ResponseTool{openai.NewResponseAsyncFunctionTool(openai.FunctionDefinition{
			Name: "lookup", Parameters: map[string]any{"type": "object"},
		})},
	})
	checks.NoError(t, err, "create response")
}

func TestGetFinalOutputTextFiltersCommentaryAndSubagents(t *testing.T) {
	var response openai.CreateResponseResponse
	checks.NoError(t, json.Unmarshal([]byte(`{"output_text":"unsafe aggregate","output":[
		{"type":"message","phase":"commentary","content":[{"type":"output_text","text":"Working"}]},
		{"type":"message","agent":{"agent_name":"/root/worker"},"phase":"final_answer",
		 "content":[{"type":"output_text","text":"Worker"}]},
		{"type":"multi_agent_call_output","content":[{"type":"output_text","text":"Tool"}]},
		{"type":"message","agent":{"agent_name":"/root"},"phase":"final_answer",
		 "content":[{"type":"output_text","text":"Final "},{"type":"refusal","refusal":"No"}]},
		{"type":"message","content":[{"type":"output_text","text":"answer"}]}
	]}`), &response), "decode response")
	response.Output = append(response.Output, make(chan int), "unknown")
	if response.GetFinalOutputText() != "Final answer" {
		t.Errorf("unexpected final answer: %q", response.GetFinalOutputText())
	}
	if response.GetOutputText() != "unsafe aggregate" {
		t.Error("existing GetOutputText behavior changed")
	}
}

func TestResponseRetrievalBetaHeaders(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("OpenAI-Beta") != "responses_multi_agent=v1,future=v1" {
			t.Errorf("unexpected beta header: %q", r.Header.Get("OpenAI-Beta"))
		}
		if r.URL.Query().Get("include") != "reasoning.encrypted_content" {
			t.Errorf("include option lost: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(w, `{"id":"resp_multi","output":[],"data":[]}`)
		checks.NoError(t, err, "write response")
	}
	server.RegisterHandler("/v1/responses/resp_multi", handler)
	server.RegisterHandler("/v1/responses/resp_multi/input_items", handler)
	betas := []string{openai.ResponseBetaMultiAgent, "future=v1"}
	include := []openai.ResponseInclude{openai.ResponseIncludeReasoningEncryptedContent}
	_, err := client.RetrieveResponse(context.Background(), "resp_multi", openai.RetrieveResponseOptions{
		Betas: betas, Include: include,
	})
	checks.NoError(t, err, "retrieve response")
	_, err = client.ListResponseInputItems(context.Background(), "resp_multi", openai.ResponseInputItemsListOptions{
		Betas: betas, Include: include,
	})
	checks.NoError(t, err, "list input items")
}
