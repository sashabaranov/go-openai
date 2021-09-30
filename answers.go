package gogpt

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type AnswerRequest struct {
	Documents       []string   `json:"documents,omitempty"`
	File            string     `json:"file,omitempty"`
	Question        string     `json:"question"`
	SearchModel     string     `json:"search_model,omitempty"`
	Model           string     `json:"model"`
	ExamplesContext string     `json:"examples_context"`
	Examples        [][]string `json:"examples"`
	MaxTokens       int        `json:"max_tokens,omitempty"`
	Stop            []string   `json:"stop,omitempty"`
	Temperature     *float64   `json:"temperature,omitempty"`
}

type AnswerResponse struct {
	Answers           []string `json:"answers"`
	Completion        string   `json:"completion"`
	Model             string   `json:"model"`
	Object            string   `json:"object"`
	SearchModel       string   `json:"search_model"`
	SelectedDocuments []struct {
		Document int    `json:"document"`
		Text     string `json:"text"`
	} `json:"selected_documents"`
}

// Search — perform a semantic search api call over a list of documents.
func (c *Client) Answers(ctx context.Context, request AnswerRequest) (response AnswerResponse, err error) {
	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", c.fullURL("/answers"), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &response)
	return
}
