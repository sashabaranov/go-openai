package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

const testDocsURL = "https://go.dev/doc/database/cancel-operations"

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// The server returns raw wire JSON, rather than marshaling SDK response types.
// Both model turns are fixtures; the callback uses the real HTTP implementation
// against a separate local MCP server. No external services or keys are used.
func TestDocsFunctionRoundTrip(t *testing.T) {
	search := `{"results":[{"url":"` + testDocsURL +
		`","excerpts":["Always defer cancel to release resources."]}]}`
	mcp := testMCPServer(t, search, false)
	defer mcp.Close()
	first, second := fixture(t, "docs_call.json"), fixture(t, "docs_answer.json")
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost || r.URL.Path != "/responses" ||
			r.Header.Get("Authorization") != "Bearer fixture-token" {
			t.Errorf("unexpected SDK request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch requests {
		case 1:
			checkToolRequest(t, body)
			_, _ = w.Write(first)
		case 2:
			checkToolOutput(t, body)
			_, _ = w.Write(second)
		default:
			t.Errorf("unexpected extra Responses request %d", requests)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	lookup := func(ctx context.Context, query string) (docSource, error) {
		if query != "Why defer cancel after context.WithTimeout?" {
			t.Errorf("callback query = %q", query)
		}
		return parallelDocs(ctx, mcp.Client(), mcp.URL, query)
	}
	answer, err := answerWithDocs(context.Background(), fixtureClient(server), "fixture", lookup)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || !strings.Contains(answer, "[Go documentation]("+testDocsURL+")") {
		t.Fatalf("requests=%d, answer=%q", requests, answer)
	}
}

func fixtureClient(server *httptest.Server) *openai.Client {
	config := openai.DefaultConfig("fixture-token")
	config.BaseURL, config.HTTPClient = server.URL, server.Client()
	return openai.NewClientWithConfig(config)
}

func checkToolRequest(t *testing.T, body map[string]any) {
	t.Helper()
	if body["input"] != docsQuestion || body["store"] != true || body["parallel_tool_calls"] != false {
		t.Errorf("unexpected initial request: %#v", body)
	}
	tools, ok := body["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("unexpected tools: %#v", body["tools"])
	}
	tool, ok := tools[0].(map[string]any)
	if !ok || tool["type"] != "function" || tool["name"] != "find_go_docs" || tool["strict"] != true {
		t.Fatalf("unexpected inline function: %#v", tools[0])
	}
	if _, nested := tool["function"]; nested {
		t.Error("Responses function must use inline fields")
	}
	parameters, ok := tool["parameters"].(map[string]any)
	if !ok || parameters["type"] != "object" || parameters["additionalProperties"] != false {
		t.Errorf("unexpected strict function schema: %#v", tool["parameters"])
	}
	choice, ok := body["tool_choice"].(map[string]any)
	if !ok || choice["type"] != "function" || choice["name"] != "find_go_docs" {
		t.Errorf("unexpected tool choice: %#v", body["tool_choice"])
	}
}

func checkToolOutput(t *testing.T, body map[string]any) {
	t.Helper()
	if body["previous_response_id"] != "resp_docs_call" || body["store"] != true {
		t.Errorf("continuation lost response state: %#v", body)
	}
	items, ok := body["input"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("unexpected continuation input: %#v", body["input"])
	}
	item, ok := items[0].(map[string]any)
	if !ok || item["type"] != "function_call_output" || item["call_id"] != "call_docs_42" {
		t.Fatalf("function output must match call_id, not output item ID: %#v", items[0])
	}
	output, ok := item["output"].(string)
	if !ok {
		t.Fatalf("function output must be a string: %#v", item["output"])
	}
	var source docSource
	if err := json.Unmarshal([]byte(output), &source); err != nil {
		t.Fatal(err)
	}
	if source.URL != testDocsURL || source.Excerpt != "Always defer cancel to release resources." {
		t.Errorf("continuation changed the returned evidence: %#v", source)
	}
}

func TestDocsCallRejectsInvalidCalls(t *testing.T) {
	for _, tc := range []struct{ name, old, replacement string }{
		{"incomplete response", `"status": "completed"`, `"status": "incomplete"`},
		{"unknown function", `"find_go_docs"`, `"other_function"`},
		{"missing call ID", `"call_docs_42"`, `""`},
		{"unfinished call", `"status": "completed",\n      "call_id"`, `"status": "in_progress",\n      "call_id"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := strings.Replace(string(fixture(t, "docs_call.json")),
				strings.ReplaceAll(tc.old, `\n`, "\n"), strings.ReplaceAll(tc.replacement, `\n`, "\n"), 1)
			var response openai.CreateResponseResponse
			if err := json.Unmarshal([]byte(data), &response); err != nil {
				t.Fatal(err)
			}
			if _, err := docsCall(response); err == nil {
				t.Fatal("accepted invalid call")
			}
		})
	}
}

func TestDocsCallAcceptsOptionalItemStatus(t *testing.T) {
	data := strings.Replace(string(fixture(t, "docs_call.json")),
		`"status": "completed",`+"\n      ", "", 1)
	var response openai.CreateResponseResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		t.Fatal(err)
	}
	if _, err := docsCall(response); err != nil {
		t.Fatalf("completed response with omitted optional item status: %v", err)
	}
}

type docsFailure struct {
	name, first, second string
	source              docSource
	lookupErr           error
	want                string
	wantRequests        int
}

func TestDocsFunctionFailures(t *testing.T) {
	boom := errors.New("lookup unavailable")
	first := string(fixture(t, "docs_call.json"))
	source := docSource{URL: testDocsURL, Excerpt: "evidence"}
	for _, tc := range []docsFailure{
		{
			name: "first API failure", first: `{"error":{"message":"denied","type":"invalid_request_error"}}`,
			want: "denied", wantRequests: 1,
		},
		{
			name: "bad arguments", first: strings.Replace(first,
				`\"query\":\"Why defer cancel after context.WithTimeout?\"`, `\"query\":42`, 1),
			want: "decode function arguments", wantRequests: 1,
		},
		{
			name: "empty query", first: strings.Replace(first, "Why defer cancel after context.WithTimeout?", "", 1),
			want: "query is empty", wantRequests: 1,
		},
		{name: "callback failure", lookupErr: boom, want: "lookup unavailable", wantRequests: 1},
		{name: "absent result", want: "no official Go documentation", wantRequests: 1},
		{
			name: "second API failure", source: source,
			second: `{"error":{"message":"limited","type":"rate_limit_error"}}`,
			want:   "limited", wantRequests: 2,
		},
		{
			name: "missing citation", source: source, second: `{"status":"completed","output_text":"Defer cancel."}`,
			want: "exact returned documentation URL", wantRequests: 2,
		},
		{
			name: "failed response", source: source,
			second: `{"status":"failed","error":{"code":"server_error","message":"failed"}}`,
			want:   "server_error", wantRequests: 2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.first == "" {
				tc.first = first
			}
			requests, err := runDocsFailure(t, tc)
			if err == nil || !strings.Contains(err.Error(), tc.want) || requests != tc.wantRequests {
				t.Fatalf("requests=%d, error=%v; want %d, %q", requests, err, tc.wantRequests, tc.want)
			}
			if tc.lookupErr != nil && !errors.Is(err, boom) {
				t.Errorf("lost callback error: %v", err)
			}
		})
	}
}

func runDocsFailure(t *testing.T, tc docsFailure) (int, error) {
	t.Helper()
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		body := tc.first
		if requests > 1 {
			body = tc.second
		}
		if strings.Contains(body, `"type":"invalid_request_error"`) ||
			strings.Contains(body, `"type":"rate_limit_error"`) {
			w.WriteHeader(http.StatusBadRequest)
		}
		_, _ = fmt.Fprint(w, body)
	}))
	defer server.Close()
	_, err := answerWithDocs(context.Background(), fixtureClient(server), "fixture",
		func(context.Context, string) (docSource, error) { return tc.source, tc.lookupErr })
	return requests, err
}
