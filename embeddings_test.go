package openai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
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
		// test embedding request with strings (simple embedding request)
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

		// test embedding request with strings
		embeddingReqStrings := EmbeddingRequestStrings{
			Input: []string{
				"The food was delicious and the waiter",
				"Other examples of embedding request",
			},
			Model: model,
		}
		marshaled, err = json.Marshal(embeddingReqStrings)
		checks.NoError(t, err, "Could not marshal embedding request")
		if !bytes.Contains(marshaled, []byte(`"model":"`+model.String()+`"`)) {
			t.Fatalf("Expected embedding request to contain model field")
		}

		// test embedding request with tokens
		embeddingReqTokens := EmbeddingRequestTokens{
			Input: [][]int{
				{464, 2057, 373, 12625, 290, 262, 46612},
				{6395, 6096, 286, 11525, 12083, 2581},
			},
			Model: model,
		}
		marshaled, err = json.Marshal(embeddingReqTokens)
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

	sampleEmbeddings := []Embedding{
		{Embedding: []float32{1.23, 4.56, 7.89}},
		{Embedding: []float32{-0.006968617, -0.0052718227, 0.011901081}},
	}

	sampleBase64Embeddings := []Base64Embedding{
		{Embedding: "pHCdP4XrkUDhevxA"},
		{Embedding: "/1jku0G/rLvA/EI8"},
	}

	server.RegisterHandler(
		"/v1/embeddings",
		func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				EncodingFormat EmbeddingEncodingFormat `json:"encoding_format"`
				User           string                  `json:"user"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)

			var resBytes []byte
			switch {
			case req.User == "invalid":
				w.WriteHeader(http.StatusBadRequest)
				return
			case req.EncodingFormat == EmbeddingEncodingFormatBase64:
				resBytes, _ = json.Marshal(EmbeddingResponseBase64{Data: sampleBase64Embeddings})
			default:
				resBytes, _ = json.Marshal(EmbeddingResponse{Data: sampleEmbeddings})
			}
			fmt.Fprintln(w, string(resBytes))
		},
	)
	// test create embeddings with strings (simple embedding request)
	res, err := client.CreateEmbeddings(context.Background(), EmbeddingRequest{})
	checks.NoError(t, err, "CreateEmbeddings error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}

	// test create embeddings with strings (simple embedding request)
	res, err = client.CreateEmbeddings(
		context.Background(),
		EmbeddingRequest{
			EncodingFormat: EmbeddingEncodingFormatBase64,
		},
	)
	checks.NoError(t, err, "CreateEmbeddings error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}

	// test create embeddings with strings
	res, err = client.CreateEmbeddings(context.Background(), EmbeddingRequestStrings{})
	checks.NoError(t, err, "CreateEmbeddings strings error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}

	// test create embeddings with tokens
	res, err = client.CreateEmbeddings(context.Background(), EmbeddingRequestTokens{})
	checks.NoError(t, err, "CreateEmbeddings tokens error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}

	// test failed sendRequest
	_, err = client.CreateEmbeddings(context.Background(), EmbeddingRequest{
		User:           "invalid",
		EncodingFormat: EmbeddingEncodingFormatBase64,
	})
	checks.HasError(t, err, "CreateEmbeddings error")
}

func TestEmbeddingResponseBase64_ToEmbeddingResponse(t *testing.T) {
	type fields struct {
		Object string
		Data   []Base64Embedding
		Model  EmbeddingModel
		Usage  Usage
	}
	tests := []struct {
		name    string
		fields  fields
		want    EmbeddingResponse
		wantErr bool
	}{
		{
			name: "test embedding response base64 to embedding response",
			fields: fields{
				Data: []Base64Embedding{
					{Embedding: "pHCdP4XrkUDhevxA"},
					{Embedding: "/1jku0G/rLvA/EI8"},
				},
			},
			want: EmbeddingResponse{
				Data: []Embedding{
					{Embedding: []float32{1.23, 4.56, 7.89}},
					{Embedding: []float32{-0.006968617, -0.0052718227, 0.011901081}},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid embedding",
			fields: fields{
				Data: []Base64Embedding{
					{
						Embedding: "----",
					},
				},
			},
			want:    EmbeddingResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &EmbeddingResponseBase64{
				Object: tt.fields.Object,
				Data:   tt.fields.Data,
				Model:  tt.fields.Model,
				Usage:  tt.fields.Usage,
			}
			got, err := r.ToEmbeddingResponse()
			if (err != nil) != tt.wantErr {
				t.Errorf("EmbeddingResponseBase64.ToEmbeddingResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EmbeddingResponseBase64.ToEmbeddingResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
