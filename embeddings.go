package openai

import (
	"context"
	"net/http"
)

// EmbeddingModel enumerates the models which can be used
// to generate Embedding vectors.

const (
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaSimilarity = "text-similarity-ada-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	BabbageSimilarity = "text-similarity-babbage-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	CurieSimilarity = "text-similarity-curie-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	DavinciSimilarity = "text-similarity-davinci-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaSearchDocument = "text-search-ada-doc-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaSearchQuery = "text-search-ada-query-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	BabbageSearchDocument = "text-search-babbage-doc-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	BabbageSearchQuery = "text-search-babbage-query-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	CurieSearchDocument = "text-search-curie-doc-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	CurieSearchQuery = "text-search-curie-query-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	DavinciSearchDocument = "text-search-davinci-doc-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	DavinciSearchQuery = "text-search-davinci-query-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaCodeSearchCode = "code-search-ada-code-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaCodeSearchText = "code-search-ada-text-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	BabbageCodeSearchCode = "code-search-babbage-code-001"
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	BabbageCodeSearchText = "code-search-babbage-text-001"
	AdaEmbeddingV2        = "text-embedding-ada-002"
)

// Embedding is a special format of data representation that can be easily utilized by machine
// learning models and algorithms. The embedding is an information dense representation of the
// semantic meaning of a piece of text. Each embedding is a vector of floating point numbers,
// such that the distance between two embeddings in the vector space is correlated with semantic similarity
// between two inputs in the original format. For example, if two texts are similar,
// then their vector representations should also be similar.
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingResponse is the response from a Create embeddings request.
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

type EmbeddingRequestConverter interface {
	// Needs to be of type EmbeddingRequestStrings or EmbeddingRequestTokens
	Convert() EmbeddingRequest
}

type EmbeddingRequest struct {
	Input any    `json:"input"`
	Model string `json:"model"`
	User  string `json:"user"`
}

func (r EmbeddingRequest) Convert() EmbeddingRequest {
	return r
}

// EmbeddingRequestStrings is the input to a create embeddings request with a slice of strings.
type EmbeddingRequestStrings struct {
	// Input is a slice of strings for which you want to generate an Embedding vector.
	// Each input must not exceed 8192 tokens in length.
	// OpenAPI suggests replacing newlines (\n) in your input with a single space, as they
	// have observed inferior results when newlines are present.
	// E.g.
	//	"The food was delicious and the waiter..."
	Input []string `json:"input"`
	// ID of the model to use. You can use the List models API to see all of your available models,
	// or see our Model overview for descriptions of them.
	Model string `json:"model"`
	// A unique identifier representing your end-user, which will help OpenAI to monitor and detect abuse.
	User string `json:"user"`
}

func (r EmbeddingRequestStrings) Convert() EmbeddingRequest {
	return EmbeddingRequest{
		Input: r.Input,
		Model: r.Model,
		User:  r.User,
	}
}

type EmbeddingRequestTokens struct {
	// Input is a slice of slices of ints ([][]int) for which you want to generate an Embedding vector.
	// Each input must not exceed 8192 tokens in length.
	// OpenAPI suggests replacing newlines (\n) in your input with a single space, as they
	// have observed inferior results when newlines are present.
	// E.g.
	//	"The food was delicious and the waiter..."
	Input [][]int `json:"input"`
	// ID of the model to use. You can use the List models API to see all of your available models,
	// or see our Model overview for descriptions of them.
	Model string `json:"model"`
	// A unique identifier representing your end-user, which will help OpenAI to monitor and detect abuse.
	User string `json:"user"`
}

func (r EmbeddingRequestTokens) Convert() EmbeddingRequest {
	return EmbeddingRequest{
		Input: r.Input,
		Model: r.Model,
		User:  r.User,
	}
}

// CreateEmbeddings returns an EmbeddingResponse which will contain an Embedding for every item in |body.Input|.
// https://beta.openai.com/docs/api-reference/embeddings/create
//
// Body should be of type EmbeddingRequestStrings for embedding strings or EmbeddingRequestTokens
// for embedding groups of text already converted to tokens.
func (c *Client) CreateEmbeddings(ctx context.Context, conv EmbeddingRequestConverter) (res EmbeddingResponse, err error) { //nolint:lll
	baseReq := conv.Convert()
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/embeddings", baseReq.Model), withBody(baseReq))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &res)

	return
}
