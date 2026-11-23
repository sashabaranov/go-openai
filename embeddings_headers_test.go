package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestCreateEmbeddingsResponseHeaders(t *testing.T) {
	for _, tc := range []struct {
		name   string
		format openai.EmbeddingEncodingFormat
	}{
		{name: "default"},
		{name: "float", format: openai.EmbeddingEncodingFormatFloat},
		{name: "base64", format: openai.EmbeddingEncodingFormatBase64},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, server, teardown := setupOpenAITestServer()
			defer teardown()

			wantData := []openai.Embedding{
				{Object: "embedding", Index: 0, Embedding: []float32{1.23, 4.56, 7.89}},
			}
			wantUsage := openai.Usage{PromptTokens: 3, TotalTokens: 3}
			wantLimits := openai.RateLimitHeaders{
				LimitRequests: 100, LimitTokens: 1000,
				RemainingRequests: 99, RemainingTokens: 997,
				ResetRequests: "1s", ResetTokens: "2s",
			}

			server.RegisterHandler("/v1/embeddings", embeddingResponseHeadersHandler(t, tc.format, wantData, wantUsage))

			response, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
				Input: []string{"test input"}, Model: openai.AdaEmbeddingV2, EncodingFormat: tc.format,
			})
			if err != nil {
				t.Fatalf("CreateEmbeddings: %v", err)
			}
			if got := response.Header().Get("X-Request-ID"); got != "req_embeddings" {
				t.Errorf("request ID = %q, want req_embeddings", got)
			}
			if got := response.Header().Values("X-Trace"); !reflect.DeepEqual(got, []string{"first", "second"}) {
				t.Errorf("trace headers = %v, want [first second]", got)
			}
			if got := response.GetRateLimitHeaders(); got != wantLimits {
				t.Errorf("rate limits = %#v, want %#v", got, wantLimits)
			}
			if response.Object != "list" || response.Model != openai.AdaEmbeddingV2 || response.Usage != wantUsage {
				t.Errorf("unexpected response metadata: %#v", response)
			}
			if !reflect.DeepEqual(response.Data, wantData) {
				t.Errorf("embedding data = %#v, want %#v", response.Data, wantData)
			}
		})
	}
}

func embeddingResponseHeadersHandler(
	t *testing.T,
	format openai.EmbeddingEncodingFormat,
	wantData []openai.Embedding,
	wantUsage openai.Usage,
) func(http.ResponseWriter, *http.Request) {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		var req openai.EmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode embedding request: %v", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if req.EncodingFormat != format {
			t.Errorf("encoding format = %q, want %q", req.EncodingFormat, format)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", "req_embeddings")
		w.Header().Add("X-Trace", "first")
		w.Header().Add("X-Trace", "second")
		w.Header().Set("X-Ratelimit-Limit-Requests", "100")
		w.Header().Set("X-Ratelimit-Limit-Tokens", "1000")
		w.Header().Set("X-Ratelimit-Remaining-Requests", "99")
		w.Header().Set("X-Ratelimit-Remaining-Tokens", "997")
		w.Header().Set("X-Ratelimit-Reset-Requests", "1s")
		w.Header().Set("X-Ratelimit-Reset-Tokens", "2s")

		var response any = openai.EmbeddingResponse{
			Object: "list", Data: wantData, Model: openai.AdaEmbeddingV2, Usage: wantUsage,
		}
		if format == openai.EmbeddingEncodingFormatBase64 {
			response = openai.EmbeddingResponseBase64{
				Object: "list", Model: openai.AdaEmbeddingV2, Usage: wantUsage,
				Data: []openai.Base64Embedding{
					{Object: "embedding", Index: 0, Embedding: "pHCdP4XrkUDhevxA"},
				},
			}
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("encode embedding response: %v", err)
		}
	}
}
