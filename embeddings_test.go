package openai_test

import (
	"errors"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestEmbedding(t *testing.T) {
	embeddedModels := []EmbeddingModel{
		AdaSimilarity,
		BabbageSimilarity,
		CurieSimilarity,
		DavinciSimilarity,
		AdaSearchDocument,
		AdaSearchQuery,
		BabbageSearchDocument,
		BabbageSearchQuery,
		CurieSearchDocument,
		CurieSearchQuery,
		DavinciSearchDocument,
		DavinciSearchQuery,
		AdaCodeSearchCode,
		AdaCodeSearchText,
		BabbageCodeSearchCode,
		BabbageCodeSearchText,
	}
	for _, model := range embeddedModels {
		embeddingReq := EmbeddingRequest{
			Input: []string{
				"The food was delicious and the waiter",
				"Other examples of embedding request",
			},
			Model: model,
		}
		// marshal embeddingReq to JSON and confirm that the model field equals
		// the AdaSearchQuery type
		marshaled, err := json.Marshal(embeddingReq)
		checks.NoError(t, err, "Could not marshal embedding request")
		if !bytes.Contains(marshaled, []byte(`"model":"`+model.String()+`"`)) {
			t.Fatalf("Expected embedding request to contain model field")
		}
	}
}

func TestEmbeddingModel(t *testing.T) {
	var em EmbeddingModel
	err := em.UnmarshalText([]byte("text-similarity-ada-001"))
	checks.NoError(t, err, "Could not marshal embedding model")

	if em != AdaSimilarity {
		t.Errorf("Model is not equal to AdaSimilarity")
	}

	err = em.UnmarshalText([]byte("some-non-existent-model"))
	checks.NoError(t, err, "Could not marshal embedding model")
	if em != Unknown {
		t.Errorf("Model is not equal to Unknown")
	}
}

func TestEmbeddingEndpoint(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler(
		"/v1/embeddings",
		func(w http.ResponseWriter, r *http.Request) {
			resBytes, _ := json.Marshal(EmbeddingResponse{})
			fmt.Fprintln(w, string(resBytes))
		},
	)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err = client.CreateEmbeddings(ctx, EmbeddingRequest{})
	checks.NoError(t, err, "CreateEmbeddings error")
}

func TestEmbeddingRateLimit(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler(
		"/v1/embeddings",
		func(w http.ResponseWriter, r *http.Request) {
			resBytes, _ := json.Marshal(EmbeddingResponse{})
			fmt.Fprintln(w, string(resBytes))
		},
	)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.EnableRateLimiter = true
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.CreateEmbeddings(ctx, EmbeddingRequest{
		Model: AdaEmbeddingV2,
	})
	checks.ErrorContains(t, err, "context canceled", "CreateEmbeddings error")
}

func TestEmbeddingRequest_Tokens(t *testing.T) {
	testcases := []struct {
		name       string
		model      EmbeddingModel
		input      []string
		wantErr    error
		wantTokens int
	}{
		{
			name:    "test unknown model",
			wantErr: errors.New("failed to tokenize prompt: model not supported: model not supported"),
		},
		{
			name:  "test1",
			model: AdaEmbeddingV2,
			input: []string{
				"The food was delicious and the waiter",
			},
			wantTokens: 7,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(tt *testing.T) {
			req := EmbeddingRequest{
				Model: testcase.model,
				Input: testcase.input,
			}
			tokens, err := req.Tokens()
			if err != nil && testcase.wantErr == nil {
				tt.Fatalf("Tokens() returned unexpected error: %v", err)
			}

			if err != nil && testcase.wantErr != nil && err.Error() != testcase.wantErr.Error() {
				tt.Fatalf("Tokens() returned unexpected error: %v, want: %v", err, testcase.wantErr)
			}

			if tokens != testcase.wantTokens {
				tt.Fatalf("Tokens() returned unexpected number of tokens: %d, want: %d", tokens, testcase.wantTokens)
			}
		})
	}
}
