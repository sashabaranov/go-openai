package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

func setup(t *testing.T, path, resp string) (*httptest.Server, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path {
			w.Header().Set("Content-Type", "application/json")
			if _, err := io.WriteString(w, resp); err != nil {
				t.Fatalf("write: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	orig := http.DefaultTransport
	http.DefaultTransport = rewriteTransport{host: strings.TrimPrefix(server.URL, "http://"), rt: orig}
	return server, func() { http.DefaultTransport = orig; server.Close() }
}

func TestMainExample(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test")
	_, cleanup := setup(t, "/v1/chat/completions", `{"choices":[{"message":{"content":"hi"}}]}`)
	defer cleanup()

	r, w, _ := os.Pipe()
	if _, err := w.Write([]byte("hello\n")); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	w.Close()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin; r.Close() }()

	main()
}
