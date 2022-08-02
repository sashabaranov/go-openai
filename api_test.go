package gogpt

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestAPI(t *testing.T) {
	tokenBytes, err := ioutil.ReadFile(".openai-token")
	if err != nil {
		t.Fatalf("Could not load auth token from .openai-token file")
	}

	c := NewClient(string(tokenBytes))
	ctx := context.Background()
	_, err = c.ListEngines(ctx)
	if err != nil {
		t.Fatalf("ListEngines error: %v", err)
	}

	_, err = c.GetEngine(ctx, "davinci")
	if err != nil {
		t.Fatalf("GetEngine error: %v", err)
	}

	fileRes, err := c.ListFiles(ctx)
	if err != nil {
		t.Fatalf("ListFiles error: %v", err)
	}

	if len(fileRes.Files) > 0 {
		_, err = c.GetFile(ctx, fileRes.Files[0].ID)
		if err != nil {
			t.Fatalf("GetFile error: %v", err)
		}
	} // else skip

	req := CompletionRequest{
		MaxTokens: 5,
		Model:     "ada",
	}
	req.Prompt = "Lorem ipsum"
	_, err = c.CreateCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateCompletion error: %v", err)
	}

	searchReq := SearchRequest{
		Documents: []string{"White House", "hospital", "school"},
		Query:     "the president",
	}
	_, err = c.Search(ctx, "ada", searchReq)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	embeddingReq := EmbeddingRequest{
		Input: []string{
			"The food was delicious and the waiter",
			"Other examples of embedding request",
		},
		Model: AdaSearchQuery,
	}
	_, err = c.CreateEmbeddings(ctx, embeddingReq)
	if err != nil {
		t.Fatalf("Embedding error: %v", err)
	}
}

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
		if err != nil {
			t.Fatalf("Could not marshal embedding request: %v", err)
		}
		if !bytes.Contains(marshaled, []byte(`"model":"`+model.String()+`"`)) {
			t.Fatalf("Expected embedding request to contain model field")
		}
	}
}
