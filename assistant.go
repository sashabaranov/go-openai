package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	assistantsSuffix      = "/assistants"
	assistantsFilesSuffix = "/files"
)

type Assistant struct {
	ID             string                 `json:"id"`
	Object         string                 `json:"object"`
	CreatedAt      int64                  `json:"created_at"`
	Name           *string                `json:"name,omitempty"`
	Description    *string                `json:"description,omitempty"`
	Model          string                 `json:"model"`
	Instructions   *string                `json:"instructions,omitempty"`
	Tools          []AssistantTool        `json:"tools"`
	ToolResources  *AssistantToolResource `json:"tool_resources,omitempty"`
	FileIDs        []string               `json:"file_ids,omitempty"` // Deprecated in v2
	Metadata       map[string]any         `json:"metadata,omitempty"`
	Temperature    *float32               `json:"temperature,omitempty"`
	TopP           *float32               `json:"top_p,omitempty"`
	ResponseFormat any                    `json:"response_format,omitempty"`

	httpHeader
}

type AssistantToolType string

const (
	AssistantToolTypeCodeInterpreter AssistantToolType = "code_interpreter"
	AssistantToolTypeRetrieval       AssistantToolType = "retrieval"
	AssistantToolTypeFunction        AssistantToolType = "function"
	AssistantToolTypeFileSearch      AssistantToolType = "file_search"
)

type AssistantTool struct {
	Type     AssistantToolType   `json:"type"`
	Function *FunctionDefinition `json:"function,omitempty"`
}

type AssistantToolFileSearch struct {
	VectorStoreIDs []string `json:"vector_store_ids"`
}

type AssistantToolCodeInterpreter struct {
	FileIDs []string `json:"file_ids"`
}

type AssistantToolResource struct {
	FileSearch      *AssistantToolFileSearch      `json:"file_search,omitempty"`
	CodeInterpreter *AssistantToolCodeInterpreter `json:"code_interpreter,omitempty"`
}

// AssistantRequest provides the assistant request parameters.
// When modifying the tools the API functions as the following:
// If Tools is undefined, no changes are made to the Assistant's tools.
// If Tools is empty slice it will effectively delete all of the Assistant's tools.
// If Tools is populated, it will replace all of the existing Assistant's tools with the provided tools.
type AssistantRequest struct {
	Model          string                 `json:"model"`
	Name           *string                `json:"name,omitempty"`
	Description    *string                `json:"description,omitempty"`
	Instructions   *string                `json:"instructions,omitempty"`
	Tools          []AssistantTool        `json:"-"`
	FileIDs        []string               `json:"file_ids,omitempty"`
	Metadata       map[string]any         `json:"metadata,omitempty"`
	ToolResources  *AssistantToolResource `json:"tool_resources,omitempty"`
	ResponseFormat any                    `json:"response_format,omitempty"`
	Temperature    *float32               `json:"temperature,omitempty"`
	TopP           *float32               `json:"top_p,omitempty"`
}

// MarshalJSON provides a custom marshaller for the assistant request to handle the API use cases
// If Tools is nil, the field is omitted from the JSON.
// If Tools is an empty slice, it's included in the JSON as an empty array ([]).
// If Tools is populated, it's included in the JSON with the elements.
func (a AssistantRequest) MarshalJSON() ([]byte, error) {
	type Alias AssistantRequest
	assistantAlias := &struct {
		Tools *[]AssistantTool `json:"tools,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&a),
	}

	if a.Tools != nil {
		assistantAlias.Tools = &a.Tools
	}

	return json.Marshal(assistantAlias)
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
		withBetaAssistantVersion(c.config.AssistantVersion))
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
		withBetaAssistantVersion(c.config.AssistantVersion))
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
		withBetaAssistantVersion(c.config.AssistantVersion))
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
		withBetaAssistantVersion(c.config.AssistantVersion))
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
) (response AssistantsList, err error) {
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
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
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
		withBetaAssistantVersion(c.config.AssistantVersion))
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
		withBetaAssistantVersion(c.config.AssistantVersion))
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
		withBetaAssistantVersion(c.config.AssistantVersion))
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
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
