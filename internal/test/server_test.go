package test

import (
	"io"
	"net/http"
	"testing"
)

func TestGetTestToken(t *testing.T) {
	if GetTestToken() != testAPI {
		t.Fatalf("unexpected token")
	}
}

func TestNewTestServer(t *testing.T) {
	ts := NewTestServer()
	if ts == nil || ts.handlers == nil {
		t.Fatalf("server not properly initialized")
	}
	if len(ts.handlers) != 0 {
		t.Fatalf("expected no handlers initially")
	}
}

func TestRegisterHandlerTransformsPath(t *testing.T) {
	ts := NewTestServer()
	h := func(w http.ResponseWriter, r *http.Request) {}
	ts.RegisterHandler("/foo/*", h)
	if ts.handlers["/foo/.*"] == nil {
		t.Fatalf("handler not registered with transformed path")
	}
}

func TestOpenAITestServer(t *testing.T) {
	ts := NewTestServer()
	ts.RegisterHandler("/v1/test/*", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	})
	srv := ts.OpenAITestServer()
	srv.Start()
	defer srv.Close()

	base := srv.Client().Transport
	client := &http.Client{Transport: &TokenRoundTripper{Token: GetTestToken(), Fallback: base}}
	resp, err := client.Get(srv.URL + "/v1/test/123")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(body) != "ok" {
		t.Fatalf("unexpected response: %d %q", resp.StatusCode, string(body))
	}

	// unregistered path
	resp, err = client.Get(srv.URL + "/unknown")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	// missing token should return unauthorized
	clientNoToken := srv.Client()
	resp, err = clientNoToken.Get(srv.URL + "/v1/test/123")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}
