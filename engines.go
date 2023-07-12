package openai

import (
	"context"
	"fmt"
	"net/http"
)

// Engine struct represents engine from OpenAPI API.
type Engine struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Owner  string `json:"owner"`
	Ready  bool   `json:"ready"`
}

// EnginesList is a list of engines.
type EnginesList struct {
	Engines []Engine `json:"data"`
}

// ListEngines Lists the currently available engines, and provides basic
// information about each option such as the owner and availability.
func (c *Client) ListEngines(ctx context.Context) (engines EnginesList, err error) {
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL("/engines"))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &engines)
	return
}

// GetEngine Retrieves an engine instance, providing basic information about
// the engine such as the owner and availability.
func (c *Client) GetEngine(
	ctx context.Context,
	engineID string,
) (engine Engine, err error) {
	urlSuffix := fmt.Sprintf("/engines/%s", engineID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &engine)
	return
}
