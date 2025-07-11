package test_test

import (
	"io"
	"net/http"
	"testing"

	testpkg "github.com/sashabaranov/go-openai/internal/test"
)

func TestOpenAITestServerAuthAndHandler(t *testing.T) {
	ts := testpkg.NewTestServer()
	ts.RegisterHandler("/v1/test", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(w, "ok"); err != nil {
			t.Fatalf("write: %v", err)
		}
	})
	srv := ts.OpenAITestServer()
	srv.Start()
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+testpkg.GetTestToken())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || string(body) != "ok" {
		t.Fatalf("unexpected response: %v %s", res.StatusCode, string(body))
	}

	req2, _ := http.NewRequest("GET", srv.URL+"/v1/test", nil)
	res2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("request2 failed: %v", err)
	}
	if res2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %v", res2.StatusCode)
	}
}

func TestOpenAITestServerWildcard(t *testing.T) {
	ts := testpkg.NewTestServer()
	ts.RegisterHandler("/v1/items/*", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := ts.OpenAITestServer()
	srv.Start()
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/v1/items/123", nil)
	req.Header.Set("api-key", testpkg.GetTestToken())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status: %v", res.StatusCode)
	}
}
