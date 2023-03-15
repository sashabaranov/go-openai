package openai

import (
	"context"
	"net/http"
)

// EmbeddingModel enumerates the models which can be used
// to generate Embedding vectors.
type EmbeddingModel int

// String implements the fmt.Stringer interface.
func (e EmbeddingModel) String() string {
	return enumToString[e]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (e EmbeddingModel) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// On unrecognized value, it sets |e| to Unknown.
func (e *EmbeddingModel) UnmarshalText(b []byte) error {
	if val, ok := stringToEnum[(string(b))]; ok {
		*e = val
		return nil
	}

	*e = Unknown

	return nil
}

const (
	Unknown EmbeddingModel = iota
	AdaSimilarity
	BabbageSimilarity
	CurieSimilarity
	DavinciSimilarity
	AdaSearchDocument
	AdaSearchQuery
	BabbageSearchDocument
	BabbageSearchQuery
	CurieSearchDocument
	CurieSearchQuery
	DavinciSearchDocument
	DavinciSearchQuery
	AdaCodeSearchCode
	AdaCodeSearchText
	BabbageCodeSearchCode
	BabbageCodeSearchText
	AdaEmbeddingV2
)

var enumToString = map[EmbeddingModel]string{
	AdaSimilarity:         "text-similarity-ada-001",
	BabbageSimilarity:     "text-similarity-babbage-001",
	CurieSimilarity:       "text-similarity-curie-001",
	DavinciSimilarity:     "text-similarity-davinci-001",
	AdaSearchDocument:     "text-search-ada-doc-001",
	AdaSearchQuery:        "text-search-ada-query-001",
	BabbageSearchDocument: "text-search-babbage-doc-001",
	BabbageSearchQuery:    "text-search-babbage-query-001",
	CurieSearchDocument:   "text-search-curie-doc-001",
	CurieSearchQuery:      "text-search-curie-query-001",
	DavinciSearchDocument: "text-search-davinci-doc-001",
	DavinciSearchQuery:    "text-search-davinci-query-001",
	AdaCodeSearchCode:     "code-search-ada-code-001",
	AdaCodeSearchText:     "code-search-ada-text-001",
	BabbageCodeSearchCode: "code-search-babbage-code-001",
	BabbageCodeSearchText: "code-search-babbage-text-001",
	AdaEmbeddingV2:        "text-embedding-ada-002",
}

var stringToEnum = map[string]EmbeddingModel{
	"text-similarity-ada-001":       AdaSimilarity,
	"text-similarity-babbage-001":   BabbageSimilarity,
	"text-similarity-curie-001":     CurieSimilarity,
	"text-similarity-davinci-001":   DavinciSimilarity,
	"text-search-ada-doc-001":       AdaSearchDocument,
	"text-search-ada-query-001":     AdaSearchQuery,
	"text-search-babbage-doc-001":   BabbageSearchDocument,
	"text-search-babbage-query-001": BabbageSearchQuery,
	"text-search-curie-doc-001":     CurieSearchDocument,
	"text-search-curie-query-001":   CurieSearchQuery,
	"text-search-davinci-doc-001":   DavinciSearchDocument,
	"text-search-davinci-query-001": DavinciSearchQuery,
	"code-search-ada-code-001":      AdaCodeSearchCode,
	"code-search-ada-text-001":      AdaCodeSearchText,
	"code-search-babbage-code-001":  BabbageCodeSearchCode,
	"code-search-babbage-text-001":  BabbageCodeSearchText,
	"text-embedding-ada-002":        AdaEmbeddingV2,
}

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
	Object string         `json:"object"`
	Data   []Embedding    `json:"data"`
	Model  EmbeddingModel `json:"model"`
	Usage  Usage          `json:"usage"`
}

// EmbeddingRequest is the input to a Create embeddings request.
type EmbeddingRequest struct {
	// Input is a slice of strings for which you want to generate an Embedding vector.
	// Each input must not exceed 2048 tokens in length.
	// OpenAPI suggests replacing newlines (\n) in your input with a single space, as they
	// have observed inferior results when newlines are present.
	// E.g.
	//	"The food was delicious and the waiter..."
	Input []string `json:"input"`
	// ID of the model to use. You can use the List models API to see all of your available models,
	// or see our Model overview for descriptions of them.
	Model EmbeddingModel `json:"model"`
	// A unique identifier representing your end-user, which will help OpenAI to monitor and detect abuse.
	User string `json:"user"`
}

// CreateEmbeddings returns an EmbeddingResponse which will contain an Embedding for every item in |request.Input|.
// https://beta.openai.com/docs/api-reference/embeddings/create
func (c *Client) CreateEmbeddings(ctx context.Context, request EmbeddingRequest) (resp EmbeddingResponse, err error) {
	req, err := c.requestBuilder.build(ctx, http.MethodPost, c.fullURL("/embeddings"), request)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &resp)

	return
}
