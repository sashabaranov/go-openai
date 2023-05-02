package openai

import (
	"context"
	"net/http"
)

// Model struct represents an OpenAPI model.
type Model struct {
	CreatedAt  int64        `json:"created"`
	ID         string       `json:"id"`
	Object     string       `json:"object"`
	OwnedBy    string       `json:"owned_by"`
	Permission []Permission `json:"permission"`
	Root       string       `json:"root"`
	Parent     string       `json:"parent"`
}

// Permission struct represents an OpenAPI permission.
type Permission struct {
	CreatedAt          int64       `json:"created"`
	ID                 string      `json:"id"`
	Object             string      `json:"object"`
	AllowCreateEngine  bool        `json:"allow_create_engine"`
	AllowSampling      bool        `json:"allow_sampling"`
	AllowLogprobs      bool        `json:"allow_logprobs"`
	AllowSearchIndices bool        `json:"allow_search_indices"`
	AllowView          bool        `json:"allow_view"`
	AllowFineTuning    bool        `json:"allow_fine_tuning"`
	Organization       string      `json:"organization"`
	Group              interface{} `json:"group"`
	IsBlocking         bool        `json:"is_blocking"`
}

// ModelsList is a list of models, including those that belong to the user or organization.
type ModelsList struct {
	Models []Model `json:"data"`
}

// ListModels Lists the currently available models,
// and provides basic information about each model such as the model id and parent.
func (c *Client) ListModels(ctx context.Context) (models ModelsList, err error) {
	req, err := c.requestBuilder.build(ctx, http.MethodGet, c.fullURL("/models"), nil)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &models)
	return
}
