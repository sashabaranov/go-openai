package test_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	internaltest "github.com/sashabaranov/go-openai/internal/test"
)

func TestCreateTestFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	internaltest.CreateTestFile(t, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected file contents: %q", string(data))
	}
}

func TestTokenRoundTripperAddsHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+internaltest.GetTestToken() {
			t.Fatalf("authorization header not set")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := srv.Client()
	client.Transport = &internaltest.TokenRoundTripper{Token: internaltest.GetTestToken(), Fallback: client.Transport}

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client request error: %v", err)
	}
	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}
