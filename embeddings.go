package openai

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"math"
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
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaSimilarity
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	BabbageSimilarity
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	CurieSimilarity
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	DavinciSimilarity
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaSearchDocument
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaSearchQuery
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	BabbageSearchDocument
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	BabbageSearchQuery
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	CurieSearchDocument
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	CurieSearchQuery
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	DavinciSearchDocument
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	DavinciSearchQuery
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaCodeSearchCode
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	AdaCodeSearchText
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
	BabbageCodeSearchCode
	// Deprecated: Will be shut down on January 04, 2024. Use text-embedding-ada-002 instead.
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

type base64String string

func (b base64String) Decode() ([]float32, error) {
	decodedData, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return nil, err
	}

	const sizeOfFloat32 = 4
	floats := make([]float32, len(decodedData)/sizeOfFloat32)
	for i := 0; i < len(floats); i++ {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(decodedData[i*4 : (i+1)*4]))
	}

	return floats, nil
}

// Base64Embedding is a container for base64 encoded embeddings.
type Base64Embedding struct {
	Object    string       `json:"object"`
	Embedding base64String `json:"embedding"`
	Index     int          `json:"index"`
}

// EmbeddingResponseBase64 is the response from a Create embeddings request with base64 encoding format.
type EmbeddingResponseBase64 struct {
	Object string            `json:"object"`
	Data   []Base64Embedding `json:"data"`
	Model  EmbeddingModel    `json:"model"`
	Usage  Usage             `json:"usage"`
}

// ToEmbeddingResponse converts an embeddingResponseBase64 to an EmbeddingResponse.
func (r *EmbeddingResponseBase64) ToEmbeddingResponse() (EmbeddingResponse, error) {
	data := make([]Embedding, len(r.Data))

	for i, base64Embedding := range r.Data {
		embedding, err := base64Embedding.Embedding.Decode()
		if err != nil {
			return EmbeddingResponse{}, err
		}

		data[i] = Embedding{
			Object:    base64Embedding.Object,
			Embedding: embedding,
			Index:     base64Embedding.Index,
		}
	}

	return EmbeddingResponse{
		Object: r.Object,
		Model:  r.Model,
		Data:   data,
		Usage:  r.Usage,
	}, nil
}

type EmbeddingRequestConverter interface {
	// Needs to be of type EmbeddingRequestStrings or EmbeddingRequestTokens
	Convert() EmbeddingRequest
}

// EmbeddingEncodingFormat is the format of the embeddings data.
// Currently, only "float" and "base64" are supported, however, "base64" is not officially documented.
// If not specified OpenAI will use "float".
type EmbeddingEncodingFormat string

const (
	EmbeddingEncodingFormatFloat  EmbeddingEncodingFormat = "float"
	EmbeddingEncodingFormatBase64 EmbeddingEncodingFormat = "base64"
)

type EmbeddingRequest struct {
	Input          any                     `json:"input"`
	Model          EmbeddingModel          `json:"model"`
	User           string                  `json:"user"`
	EncodingFormat EmbeddingEncodingFormat `json:"encoding_format,omitempty"`
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
	Model EmbeddingModel `json:"model"`
	// A unique identifier representing your end-user, which will help OpenAI to monitor and detect abuse.
	User string `json:"user"`
	// EmbeddingEncodingFormat is the format of the embeddings data.
	// Currently, only "float" and "base64" are supported, however, "base64" is not officially documented.
	// If not specified OpenAI will use "float".
	EncodingFormat EmbeddingEncodingFormat `json:"encoding_format,omitempty"`
}

func (r EmbeddingRequestStrings) Convert() EmbeddingRequest {
	return EmbeddingRequest{
		Input:          r.Input,
		Model:          r.Model,
		User:           r.User,
		EncodingFormat: r.EncodingFormat,
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
	Model EmbeddingModel `json:"model"`
	// A unique identifier representing your end-user, which will help OpenAI to monitor and detect abuse.
	User string `json:"user"`
	// EmbeddingEncodingFormat is the format of the embeddings data.
	// Currently, only "float" and "base64" are supported, however, "base64" is not officially documented.
	// If not specified OpenAI will use "float".
	EncodingFormat EmbeddingEncodingFormat `json:"encoding_format,omitempty"`
}

func (r EmbeddingRequestTokens) Convert() EmbeddingRequest {
	return EmbeddingRequest{
		Input:          r.Input,
		Model:          r.Model,
		User:           r.User,
		EncodingFormat: r.EncodingFormat,
	}
}

// CreateEmbeddings returns an EmbeddingResponse which will contain an Embedding for every item in |body.Input|.
// https://beta.openai.com/docs/api-reference/embeddings/create
//
// Body should be of type EmbeddingRequestStrings for embedding strings or EmbeddingRequestTokens
// for embedding groups of text already converted to tokens.
func (c *Client) CreateEmbeddings(
	ctx context.Context,
	conv EmbeddingRequestConverter,
) (res EmbeddingResponse, err error) {
	baseReq := conv.Convert()
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/embeddings", baseReq.Model.String()), withBody(baseReq))
	if err != nil {
		return
	}

	if baseReq.EncodingFormat != EmbeddingEncodingFormatBase64 {
		err = c.sendRequest(req, &res)
		return
	}

	base64Response := &EmbeddingResponseBase64{}
	err = c.sendRequest(req, base64Response)
	if err != nil {
		return
	}

	res, err = base64Response.ToEmbeddingResponse()
	return
}
