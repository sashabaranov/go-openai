package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	vectorStoresSuffix            = "/vector_stores"
	vectorStoresFilesSuffix       = "/files"
	vectorStoresFileBatchesSuffix = "/file_batches"
)

type VectorStoreFileCount struct {
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	Cancelled  int `json:"cancelled"`
	Total      int `json:"total"`
}

type VectorStore struct {
	ID           string               `json:"id"`
	Object       string               `json:"object"`
	CreatedAt    int64                `json:"created_at"`
	Name         string               `json:"name"`
	UsageBytes   int                  `json:"usage_bytes"`
	FileCounts   VectorStoreFileCount `json:"file_counts"`
	Status       string               `json:"status"`
	ExpiresAfter *VectorStoreExpires  `json:"expires_after"`
	ExpiresAt    *int                 `json:"expires_at"`
	Metadata     map[string]any       `json:"metadata"`

	httpHeader
}

type VectorStoreExpires struct {
	Anchor string `json:"anchor"`
	Days   int    `json:"days"`
}

// VectorStoreRequest provides the vector store request parameters.
type VectorStoreRequest struct {
	Name         string              `json:"name,omitempty"`
	FileIDs      []string            `json:"file_ids,omitempty"`
	ExpiresAfter *VectorStoreExpires `json:"expires_after,omitempty"`
	Metadata     map[string]any      `json:"metadata,omitempty"`
}

// VectorStoresList is a list of vector store.
type VectorStoresList struct {
	VectorStores []VectorStore `json:"data"`
	LastID       *string       `json:"last_id"`
	FirstID      *string       `json:"first_id"`
	HasMore      bool          `json:"has_more"`
	httpHeader
}

type VectorStoreDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

type VectorStoreFile struct {
	ID            string `json:"id"`
	Object        string `json:"object"`
	CreatedAt     int64  `json:"created_at"`
	VectorStoreID string `json:"vector_store_id"`
	UsageBytes    int    `json:"usage_bytes"`
	Status        string `json:"status"`

	httpHeader
}

type VectorStoreFileRequest struct {
	FileID string `json:"file_id"`
}

type VectorStoreFilesList struct {
	VectorStoreFiles []VectorStoreFile `json:"data"`
	FirstID          *string           `json:"first_id"`
	LastID           *string           `json:"last_id"`
	HasMore          bool              `json:"has_more"`

	httpHeader
}

type VectorStoreFileBatch struct {
	ID            string               `json:"id"`
	Object        string               `json:"object"`
	CreatedAt     int64                `json:"created_at"`
	VectorStoreID string               `json:"vector_store_id"`
	Status        string               `json:"status"`
	FileCounts    VectorStoreFileCount `json:"file_counts"`

	httpHeader
}

type VectorStoreFileBatchRequest struct {
	FileIDs []string `json:"file_ids"`
}

// CreateVectorStore creates a new vector store.
func (c *Client) CreateVectorStore(ctx context.Context, request VectorStoreRequest) (response VectorStore, err error) {
	req, _ := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(vectorStoresSuffix),
		withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion),
	)

	err = c.sendRequest(req, &response)
	return
}

// RetrieveVectorStore retrieves an vector store.
func (c *Client) RetrieveVectorStore(
	ctx context.Context,
	vectorStoreID string,
) (response VectorStore, err error) {
	urlSuffix := fmt.Sprintf("%s/%s", vectorStoresSuffix, vectorStoreID)
	req, _ := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// ModifyVectorStore modifies a vector store.
func (c *Client) ModifyVectorStore(
	ctx context.Context,
	vectorStoreID string,
	request VectorStoreRequest,
) (response VectorStore, err error) {
	urlSuffix := fmt.Sprintf("%s/%s", vectorStoresSuffix, vectorStoreID)
	req, _ := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// DeleteVectorStore deletes an vector store.
func (c *Client) DeleteVectorStore(
	ctx context.Context,
	vectorStoreID string,
) (response VectorStoreDeleteResponse, err error) {
	urlSuffix := fmt.Sprintf("%s/%s", vectorStoresSuffix, vectorStoreID)
	req, _ := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// ListVectorStores Lists the currently available vector store.
func (c *Client) ListVectorStores(
	ctx context.Context,
	pagination Pagination,
) (response VectorStoresList, err error) {
	urlValues := url.Values{}

	if pagination.After != nil {
		urlValues.Add("after", *pagination.After)
	}
	if pagination.Order != nil {
		urlValues.Add("order", *pagination.Order)
	}
	if pagination.Limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *pagination.Limit))
	}
	if pagination.Before != nil {
		urlValues.Add("before", *pagination.Before)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("%s%s", vectorStoresSuffix, encodedValues)
	req, _ := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// CreateVectorStoreFile creates a new vector store file.
func (c *Client) CreateVectorStoreFile(
	ctx context.Context,
	vectorStoreID string,
	request VectorStoreFileRequest,
) (response VectorStoreFile, err error) {
	urlSuffix := fmt.Sprintf("%s/%s%s", vectorStoresSuffix, vectorStoreID, vectorStoresFilesSuffix)
	req, _ := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// RetrieveVectorStoreFile retrieves a vector store file.
func (c *Client) RetrieveVectorStoreFile(
	ctx context.Context,
	vectorStoreID string,
	fileID string,
) (response VectorStoreFile, err error) {
	urlSuffix := fmt.Sprintf("%s/%s%s/%s", vectorStoresSuffix, vectorStoreID, vectorStoresFilesSuffix, fileID)
	req, _ := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// DeleteVectorStoreFile deletes an existing file.
func (c *Client) DeleteVectorStoreFile(
	ctx context.Context,
	vectorStoreID string,
	fileID string,
) (err error) {
	urlSuffix := fmt.Sprintf("%s/%s%s/%s", vectorStoresSuffix, vectorStoreID, vectorStoresFilesSuffix, fileID)
	req, _ := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, nil)
	return
}

// ListVectorStoreFiles Lists the currently available files for a vector store.
func (c *Client) ListVectorStoreFiles(
	ctx context.Context,
	vectorStoreID string,
	pagination Pagination,
) (response VectorStoreFilesList, err error) {
	urlValues := url.Values{}
	if pagination.After != nil {
		urlValues.Add("after", *pagination.After)
	}
	if pagination.Limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *pagination.Limit))
	}
	if pagination.Before != nil {
		urlValues.Add("before", *pagination.Before)
	}
	if pagination.Order != nil {
		urlValues.Add("order", *pagination.Order)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("%s/%s%s%s", vectorStoresSuffix, vectorStoreID, vectorStoresFilesSuffix, encodedValues)
	req, _ := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// CreateVectorStoreFileBatch creates a new vector store file batch.
func (c *Client) CreateVectorStoreFileBatch(
	ctx context.Context,
	vectorStoreID string,
	request VectorStoreFileBatchRequest,
) (response VectorStoreFileBatch, err error) {
	urlSuffix := fmt.Sprintf("%s/%s%s", vectorStoresSuffix, vectorStoreID, vectorStoresFileBatchesSuffix)
	req, _ := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// RetrieveVectorStoreFileBatch retrieves a vector store file batch.
func (c *Client) RetrieveVectorStoreFileBatch(
	ctx context.Context,
	vectorStoreID string,
	batchID string,
) (response VectorStoreFileBatch, err error) {
	urlSuffix := fmt.Sprintf("%s/%s%s/%s", vectorStoresSuffix, vectorStoreID, vectorStoresFileBatchesSuffix, batchID)
	req, _ := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// CancelVectorStoreFileBatch cancel a new vector store file batch.
func (c *Client) CancelVectorStoreFileBatch(
	ctx context.Context,
	vectorStoreID string,
	batchID string,
) (response VectorStoreFileBatch, err error) {
	urlSuffix := fmt.Sprintf("%s/%s%s/%s%s", vectorStoresSuffix,
		vectorStoreID, vectorStoresFileBatchesSuffix, batchID, "/cancel")
	req, _ := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}

// ListVectorStoreFiles Lists the currently available files for a vector store.
func (c *Client) ListVectorStoreFilesInBatch(
	ctx context.Context,
	vectorStoreID string,
	batchID string,
	pagination Pagination,
) (response VectorStoreFilesList, err error) {
	urlValues := url.Values{}
	if pagination.After != nil {
		urlValues.Add("after", *pagination.After)
	}
	if pagination.Limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *pagination.Limit))
	}
	if pagination.Before != nil {
		urlValues.Add("before", *pagination.Before)
	}
	if pagination.Order != nil {
		urlValues.Add("order", *pagination.Order)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("%s/%s%s/%s%s%s", vectorStoresSuffix,
		vectorStoreID, vectorStoresFileBatchesSuffix, batchID, "/files", encodedValues)
	req, _ := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))

	err = c.sendRequest(req, &response)
	return
}
