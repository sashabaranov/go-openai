package gogpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

/*
- SearchRequest represents a request structure for search API.

- Info (*):
- 1*) You should specify either 'documents' or a 'file', but not both.
- 2*) This flag only takes effect when file is set.
*/
type SearchRequest struct {
	Query          string   `json:"query"`
	Documents      []string `json:"documents"`            // 1*
	FileID         string   `json:"file,omitempty"`       // 1*
	MaxRerank      int      `json:"max_rerank,omitempty"` // 2*
	ReturnMetadata bool     `json:"return_metadata,omitempty"`
	User           string   `json:"user,omitempty"`
}

// SearchResult represents single result from search API.
type SearchResult struct {
	Document int     `json:"document"`
	Object   string  `json:"object"`
	Score    float32 `json:"score"`
	Metadata string  `json:"metadata"` // 2*
}

// SearchResponse represents a response structure for search API.
type SearchResponse struct {
	SearchResults []SearchResult `json:"data"`
	Object        string         `json:"object"`
}

// Search — perform a semantic search api call over a list of documents.
func (c *Client) Search(
	ctx context.Context,
	engineID string,
	request SearchRequest,
) (response SearchResponse, err error) {
	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	urlSuffix := fmt.Sprintf("/engines/%s/search", engineID)
	req, err := http.NewRequest("POST", c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &response)
	return
}
