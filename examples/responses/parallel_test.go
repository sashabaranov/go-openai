package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// testMCPServer checks protocol headers, session reuse and credential separation.
func testMCPServer(t *testing.T, search string, toolError bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkMCPHeaders(t, r)
		var body struct {
			ID     int            `json:"id"`
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		if body.Method != "initialize" &&
			(r.Header.Get("Mcp-Session-Id") != "docs-session" ||
				r.Header.Get("MCP-Protocol-Version") != mcpVersion) {
			t.Error("MCP request lost session or protocol version")
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.Method {
		case "initialize":
			if body.Params["protocolVersion"] != mcpVersion {
				t.Error("unexpected initialization version")
			}
			w.Header().Set("Mcp-Session-Id", "docs-session")
			_, _ = fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-03-26"}}`)
		case "notifications/initialized":
			if body.ID != 0 {
				t.Error("notification must not have an ID")
			}
			w.WriteHeader(http.StatusAccepted)
		case "tools/call":
			checkMCPSearch(t, body.Params)
			w.Header().Set("Content-Type", "text/event-stream")
			format := "event: message\r\ndata: {\"jsonrpc\":\"2.0\",\"id\":2," +
				"\"result\":{\"isError\":%t,\"structuredContent\":%s}}\r\n\r\n"
			_, _ = fmt.Fprintf(w, format, toolError, search)
		default:
			t.Errorf("unexpected MCP method %q", body.Method)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

func TestParallelDocsAbsentAndToolError(t *testing.T) {
	for _, tc := range []struct {
		name      string
		toolError bool
		want      string
	}{
		{"absent", false, "no official HTTPS Go documentation"},
		{"tool error", true, "tool error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := testMCPServer(t, `{"results":[]}`, tc.toolError)
			defer server.Close()
			_, err := parallelDocs(context.Background(), server.Client(), server.URL, "context.WithTimeout")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v; want %q", err, tc.want)
			}
		})
	}
}

func TestGoDocResult(t *testing.T) {
	for _, raw := range []string{
		`{"results":[]}`, `not json`,
		`{"results":[{"url":"https://go.dev.evil.test/doc","excerpts":["untrusted"]}]}`,
		`{"results":[{"url":"https://user@go.dev/doc","excerpts":["untrusted"]}]}`,
		`{"results":[{"url":"http://pkg.go.dev/context","excerpts":["untrusted"]}]}`,
		`{"results":[{"url":"https://go.dev:444/doc","excerpts":["untrusted"]}]}`,
		`{"results":[{"url":"https://go.dev/doc","excerpts":[]}]}`,
	} {
		if _, err := goDocResult([]byte(raw)); err == nil {
			t.Errorf("accepted absent or invalid source: %s", raw)
		}
	}
}

func TestMCPFailures(t *testing.T) {
	for _, tc := range []struct {
		name, body, want string
		status           int
	}{
		{name: "rate limit", status: http.StatusTooManyRequests, want: "HTTP 429"},
		{name: "unauthorized", status: http.StatusUnauthorized, want: "HTTP 401"},
		{name: "RPC error",
			body: `{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"bad arguments"}}`,
			want: "MCP error -32602"},
		{name: "mismatched ID", body: `{"jsonrpc":"2.0","id":99,"result":{}}`, want: "ID mismatch"},
		{name: "malformed", body: `broken`, want: "decode MCP response"},
		{name: "oversize", body: strings.Repeat("x", maxMCPBody+1), want: "size limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if tc.status != 0 {
					w.WriteHeader(tc.status)
				}
				_, _ = fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			_, err := parallelDocs(context.Background(), server.Client(), server.URL, "context.WithTimeout")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v; want %q", err, tc.want)
			}
		})
	}
}

func TestMCPCancellationAndTimeout(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer server.Close()
	defer close(release)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := parallelDocs(ctx, server.Client(), server.URL, "context.WithTimeout")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation was not preserved: %v", err)
	}
	client := &http.Client{Timeout: 20 * time.Millisecond}
	_, err = parallelDocs(context.Background(), client, server.URL, "context.WithTimeout")
	var timeout net.Error
	if !errors.As(err, &timeout) || !timeout.Timeout() {
		t.Fatalf("timeout was not preserved: %v", err)
	}
}

type docsTransport struct {
	t     *testing.T
	calls int
}

func (transport *docsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport.calls++
	if req.Header.Get("User-Agent") != docsUserAgent ||
		req.Header.Get("Authorization") != "" || req.Header.Get("x-api-key") != "" {
		transport.t.Error("live Parallel request must have the example User-Agent and no credentials")
	}
	return http.DefaultTransport.RoundTrip(req)
}

// This opt-in check exercises real anonymous retrieval only. It does not contact
// OpenAI and does not establish that a live model completed the tool round trip.
func TestLiveParallelDocs(t *testing.T) {
	if os.Getenv("GO_OPENAI_LIVE_DOCS") != "1" {
		t.Skip("set GO_OPENAI_LIVE_DOCS=1 to run anonymous public documentation retrieval")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	transport := &docsTransport{t: t}
	source, err := parallelDocs(ctx, &http.Client{Transport: transport, Timeout: 30 * time.Second}, parallelMCPURL,
		"Why should I defer cancel after context.WithTimeout? Find official Go documentation.")
	if err != nil {
		t.Fatal(err)
	}
	if transport.calls != 3 || !officialGoURL(source.URL) || source.Excerpt == "" {
		t.Fatalf("unexpected live retrieval: calls=%d, source=%#v", transport.calls, source)
	}
	// Verify reachability of the exact returned URL without changing its scheme,
	// path or fragment. Do not forward the MCP session to the documentation site.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("returned documentation URL: HTTP %d", response.StatusCode)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	t.Logf("Anonymous retrieval only: exact source URL %s; %d MCP requests used %s, no credentials; source HTTP %d",
		source.URL, transport.calls, docsUserAgent, response.StatusCode)
}

func checkMCPHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Header.Get("User-Agent") != docsUserAgent ||
		r.Header.Get("Authorization") != "" || r.Header.Get("x-api-key") != "" {
		t.Error("MCP request must use the example User-Agent and no credentials")
	}
	if r.Method != http.MethodPost || r.Header.Get("Accept") != "application/json, text/event-stream" {
		t.Error("unexpected MCP method or Accept header")
	}
}

func checkMCPSearch(t *testing.T, params map[string]any) {
	t.Helper()
	if params["name"] != "web_search" {
		t.Error("unexpected MCP tool")
	}
	arguments, ok := params["arguments"].(map[string]any)
	if !ok || !strings.Contains(fmt.Sprint(arguments["objective"]), "context.WithTimeout") {
		t.Error("missing search objective")
	}
	if _, ok = arguments["search_queries"].([]any); !ok {
		t.Error("missing keyword queries")
	}
}
