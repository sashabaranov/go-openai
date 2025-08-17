package test_test

import (
	"io"
	"net/http"
	"testing"

	internaltest "github.com/sashabaranov/go-openai/internal/test"
)

func TestGetTestToken(t *testing.T) {
	if internaltest.GetTestToken() != "this-is-my-secure-token-do-not-steal!!" {
		t.Fatalf("unexpected token")
	}
}

func TestNewTestServer(t *testing.T) {
	ts := internaltest.NewTestServer()
	if ts == nil {
		t.Fatalf("server not properly initialized")
	}
	if ts.HandlerCount() != 0 {
		t.Fatalf("expected no handlers initially")
	}
}

func TestRegisterHandlerTransformsPath(t *testing.T) {
	ts := internaltest.NewTestServer()
	h := func(_ http.ResponseWriter, _ *http.Request) {}
	ts.RegisterHandler("/foo/*", h)
	if !ts.HasHandler("/foo/*") {
		t.Fatalf("handler not registered with transformed path")
	}
}

func TestOpenAITestServer(t *testing.T) {
	ts := internaltest.NewTestServer()
	ts.RegisterHandler("/v1/test/*", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(w, "ok"); err != nil {
			t.Fatalf("write: %v", err)
		}
	})
	srv := ts.OpenAITestServer()
	srv.Start()
	defer srv.Close()

	base := srv.Client().Transport
	client := &http.Client{Transport: &internaltest.TokenRoundTripper{Token: internaltest.GetTestToken(), Fallback: base}}
	resp, err := client.Get(srv.URL + "/v1/test/123")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
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
