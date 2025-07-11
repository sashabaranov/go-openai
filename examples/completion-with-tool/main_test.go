package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type rewriteTransport struct {
	host string
	rt   http.RoundTripper
}

func (t rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Scheme = "http"
	r.URL.Host = t.host
	return t.rt.RoundTrip(r)
}

func setupToolServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var err error
		if count == 0 {
			_, err = io.WriteString(w,
				`{"choices":[{"message":{"tool_calls":[{"id":"1","function":{"name":"get_current_weather","arguments":"{}"}}]}}]}`)
		} else {
			_, err = io.WriteString(w, `{"choices":[{"message":{"content":"done"}}]}`)
		}
		if err != nil {
			t.Fatalf("write: %v", err)
		}
		count++
	}))
	orig := http.DefaultTransport
	http.DefaultTransport = rewriteTransport{host: strings.TrimPrefix(server.URL, "http://"), rt: orig}
	return server, func() { http.DefaultTransport = orig; server.Close() }
}

func TestMainExample(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test")
	_, cleanup := setupToolServer(t)
	defer cleanup()
	main()
}
