package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	assistantsPath      = "/assistants"
	assistantsFilesPath = "/files"
)

type Assistant struct {
	ID           string  `json:"id"`
	Object       string  `json:"object"`
	CreatedAt    int64   `json:"created_at"`
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	Model        string  `json:"model"`
	Instructions *string `json:"instructions,omitempty"`
	Tools        []any   `json:"tools,omitempty"`

	httpHeader
}

type AssistantTool struct {
	Type string `json:"type"`
}

type AssistantToolCodeInterpreter struct {
	AssistantTool
}

type AssistantToolRetrieval struct {
	AssistantTool
}

type AssistantToolFunction struct {
	AssistantTool
	Function FunctionDefinition `json:"function"`
}

type AssistantRequest struct {
	Model        string         `json:"model"`
	Name         *string        `json:"name,omitempty"`
	Description  *string        `json:"description,omitempty"`
	Instructions *string        `json:"instructions,omitempty"`
	Tools        []any          `json:"tools,omitempty"`
	FileIDs      []string       `json:"file_ids,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// AssistantsList is a list of assistants.
type AssistantsList struct {
	Assistants []Assistant `json:"data"`

	httpHeader
}

type AssistantFile struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	CreatedAt   int64  `json:"created_at"`
	AssistantID string `json:"assistant_id"`

	httpHeader
}

type AssistantFileRequest struct {
	FileID string `json:"file_id"`
}

type AssistantFilesList struct {
	AssistantFiles []AssistantFile `json:"data"`

	httpHeader
}

// CreateAssistant creates a new assistant.
func (c *Client) CreateAssistant(ctx context.Context, request AssistantRequest) (response Assistant, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(assistantsPath), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveAssistant retrieves an assistant.
func (c *Client) RetrieveAssistant(
	ctx context.Context,
	assistantID string,
) (response Assistant, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(assistantsPath+"/"+assistantID))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ModifyAssistant modifies an assistant.
func (c *Client) ModifyAssistant(
	ctx context.Context,
	assistantID string,
	request AssistantRequest,
) (response Assistant, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(assistantsPath+"/"+assistantID), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// DeleteAssistant deletes an assistant.
func (c *Client) DeleteAssistant(
	ctx context.Context,
	assistantID string,
) (response Assistant, err error) {
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(assistantsPath+"/"+assistantID))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ListFiles Lists the currently available files,
// and provides basic information about each file such as the file name and purpose.
func (c *Client) ListAssistants(
	ctx context.Context,
	limit *int,
	order *string,
	after *string,
	before *string,
) (reponse AssistantsList, err error) {
	urlValues := url.Values{}
	if limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if order != nil {
		urlValues.Add("order", *order)
	}
	if after != nil {
		urlValues.Add("after", *after)
	}
	if before != nil {
		urlValues.Add("before", *before)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(assistantsPath+encodedValues))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &reponse)
	return
}

func (c *Client) CreateAssistantFile(
	ctx context.Context,
	assistantID string,
	request AssistantFileRequest,
) (response AssistantFile, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(assistantsPath+"/"+assistantID+assistantsPath),
		withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

func (c *Client) RetrieveAssistantFile(
	ctx context.Context,
	assistantID string,
	fileID string,
) (response AssistantFile, err error) {
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(assistantsPath+"/"+
		assistantID+assistantsFilesPath+"/"+fileID))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

func (c *Client) DeleteAssistantFile(
	ctx context.Context,
	assistantID string,
	fileID string,
) (err error) {
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(assistantsPath+"/"+
		assistantID+assistantsFilesPath+"/"+fileID))
	if err != nil {
		return
	}

	err = c.sendRequest(req, nil)
	return
}

func (c *Client) ListAssistantFiles(
	ctx context.Context,
	assistantID string,
	limit *int,
	order *string,
	after *string,
	before *string,
) (response AssistantFilesList, err error) {
	urlValues := url.Values{}
	if limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if order != nil {
		urlValues.Add("order", *order)
	}
	if after != nil {
		urlValues.Add("after", *after)
	}
	if before != nil {
		urlValues.Add("before", *before)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(assistantsPath+"/"+
		assistantID+"/files"+encodedValues))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
