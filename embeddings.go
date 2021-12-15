package gogpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
)

var enumToString = map[EmbeddingModel]string{
	AdaSimilarity:         "ada-similarity",
	BabbageSimilarity:     "babbage-similarity",
	CurieSimilarity:       "curie-similarity",
	DavinciSimilarity:     "davinci-similarity",
	AdaSearchDocument:     "ada-search-document",
	AdaSearchQuery:        "ada-search-query",
	BabbageSearchDocument: "babbage-search-document",
	BabbageSearchQuery:    "babbage-search-query",
	CurieSearchDocument:   "curie-search-document",
	CurieSearchQuery:      "curie-search-query",
	DavinciSearchDocument: "davinci-search-document",
	DavinciSearchQuery:    "davinci-search-query",
	AdaCodeSearchCode:     "ada-code-search-code",
	AdaCodeSearchText:     "ada-code-search-text",
	BabbageCodeSearchCode: "babbage-code-search-code",
	BabbageCodeSearchText: "babbage-code-search-text",
}

var stringToEnum = map[string]EmbeddingModel{
	"ada-similarity":           AdaSimilarity,
	"babbage-similarity":       BabbageSimilarity,
	"curie-similarity":         CurieSimilarity,
	"davinci-similarity":       DavinciSimilarity,
	"ada-search-document":      AdaSearchDocument,
	"ada-search-query":         AdaSearchQuery,
	"babbage-search-document":  BabbageSearchDocument,
	"babbage-search-query":     BabbageSearchQuery,
	"curie-search-document":    CurieSearchDocument,
	"curie-search-query":       CurieSearchQuery,
	"davinci-search-document":  DavinciSearchDocument,
	"davinci-search-query":     DavinciSearchQuery,
	"ada-code-search-code":     AdaCodeSearchCode,
	"ada-code-search-text":     AdaCodeSearchText,
	"babbage-code-search-code": BabbageCodeSearchCode,
	"babbage-code-search-text": BabbageCodeSearchText,
}

// Embedding is a special format of data representation that can be easily utilized by machine learning models and algorithms.
// The embedding is an information dense representation of the semantic meaning of a piece of text. Each embedding is a vector of
// floating point numbers, such that the distance between two embeddings in the vector space is correlated with semantic similarity
// between two inputs in the original format. For example, if two texts are similar, then their vector representations should
// also be similar.
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingResponse is the response from a Create embeddings request.
type EmbeddingResponse struct {
	Object string         `json:"object"`
	Data   []Embedding    `json:"data"`
	Model  EmbeddingModel `json:"model"`
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
}

// CreateEmbeddings returns an EmbeddingResponse which will contain an Embedding for every item in |request.Input|.
// https://beta.openai.com/docs/api-reference/embeddings/create
func (c *Client) CreateEmbeddings(ctx context.Context, request EmbeddingRequest, model EmbeddingModel) (resp EmbeddingResponse, err error) {
	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	urlSuffix := fmt.Sprintf("/engines/%s/embeddings", model)
	req, err := http.NewRequest(http.MethodPost, c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &resp)

	return
}
