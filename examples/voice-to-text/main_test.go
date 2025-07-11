package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	_, cleanup := setup(t, "/v1/audio/transcriptions", `{"text":"ok"}`)
	defer cleanup()
	path := filepath.Join(t.TempDir(), "f.mp3")
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	oldArgs := os.Args
	os.Args = []string{"cmd", path}
	defer func() { os.Args = oldArgs }()
	main()
}
