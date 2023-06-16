package openai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
	"math"
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
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler(
		"/v1/embeddings",
		func(w http.ResponseWriter, r *http.Request) {
			resBytes, _ := json.Marshal(EmbeddingResponse{})
			fmt.Fprintln(w, string(resBytes))
		},
	)
	_, err := client.CreateEmbeddings(context.Background(), EmbeddingRequest{})
	checks.NoError(t, err, "CreateEmbeddings error")
}

func TestDotProduct(t *testing.T) {
	v1 := []float32{1, 2, 3}
	v2 := []float32{2, 4, 6}
	expected := float32(28.0)
	result := DotProduct(v1, v2)
	if math.Abs(float64(result-expected)) > 1e-12 {
		t.Errorf("Unexpected result. Expected: %v, but got %v", expected, result)
	}

	v1 = []float32{1, 0, 0}
	v2 = []float32{0, 1, 0}
	expected = float32(0.0)
	result = DotProduct(v1, v2)
	if math.Abs(float64(result-expected)) > 1e-12 {
		t.Errorf("Unexpected result. Expected: %v, but got %v", expected, result)
	}
}
