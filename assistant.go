package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	assistantsSuffix      = "/assistants"
	assistantsFilesSuffix = "/files"
	openaiAssistantsV1    = "assistants=v1"
)

type Assistant struct {
	ID           string          `json:"id"`
	Object       string          `json:"object"`
	CreatedAt    int64           `json:"created_at"`
	Name         *string         `json:"name,omitempty"`
	Description  *string         `json:"description,omitempty"`
	Model        string          `json:"model"`
	Instructions *string         `json:"instructions,omitempty"`
	Tools        []AssistantTool `json:"tools,omitempty"`
	FileIDs      []string        `json:"file_ids,omitempty"`
	Metadata     map[string]any  `json:"metadata,omitempty"`

	httpHeader
}

type AssistantToolType string

const (
	AssistantToolTypeCodeInterpreter AssistantToolType = "code_interpreter"
	AssistantToolTypeRetrieval       AssistantToolType = "retrieval"
	AssistantToolTypeFunction        AssistantToolType = "function"
)

type AssistantTool struct {
	Type     AssistantToolType   `json:"type"`
	Function *FunctionDefinition `json:"function,omitempty"`
}

type AssistantRequest struct {
	Model        string          `json:"model"`
	Name         *string         `json:"name,omitempty"`
	Description  *string         `json:"description,omitempty"`
	Instructions *string         `json:"instructions,omitempty"`
	Tools        []AssistantTool `json:"tools,omitempty"`
	FileIDs      []string        `json:"file_ids,omitempty"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
}

// AssistantsList is a list of assistants.
type AssistantsList struct {
	Assistants []Assistant `json:"data"`
	LastID     *string     `json:"last_id"`
	FirstID    *string     `json:"first_id"`
	HasMore    bool        `json:"has_more"`
	httpHeader
}

type AssistantDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

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
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(assistantsSuffix), withBody(request),
		withBetaAssistantV1())
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
	urlSuffix := fmt.Sprintf("%s/%s", assistantsSuffix, assistantID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantV1())
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
	urlSuffix := fmt.Sprintf("%s/%s", assistantsSuffix, assistantID)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantV1())
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
) (response AssistantDeleteResponse, err error) {
	urlSuffix := fmt.Sprintf("%s/%s", assistantsSuffix, assistantID)
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ListAssistants Lists the currently available assistants.
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

	urlSuffix := fmt.Sprintf("%s%s", assistantsSuffix, encodedValues)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &reponse)
	return
}

// CreateAssistantFile creates a new assistant file.
func (c *Client) CreateAssistantFile(
	ctx context.Context,
	assistantID string,
	request AssistantFileRequest,
) (response AssistantFile, err error) {
	urlSuffix := fmt.Sprintf("%s/%s%s", assistantsSuffix, assistantID, assistantsFilesSuffix)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveAssistantFile retrieves an assistant file.
func (c *Client) RetrieveAssistantFile(
	ctx context.Context,
	assistantID string,
	fileID string,
) (response AssistantFile, err error) {
	urlSuffix := fmt.Sprintf("%s/%s%s/%s", assistantsSuffix, assistantID, assistantsFilesSuffix, fileID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// DeleteAssistantFile deletes an existing file.
func (c *Client) DeleteAssistantFile(
	ctx context.Context,
	assistantID string,
	fileID string,
) (err error) {
	urlSuffix := fmt.Sprintf("%s/%s%s/%s", assistantsSuffix, assistantID, assistantsFilesSuffix, fileID)
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, nil)
	return
}

// ListAssistantFiles Lists the currently available files for an assistant.
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

	urlSuffix := fmt.Sprintf("%s/%s%s%s", assistantsSuffix, assistantID, assistantsFilesSuffix, encodedValues)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantV1())
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
