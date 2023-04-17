package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// Client is OpenAI GPT-3 API client.
type Client struct {
	config ClientConfig

	requestBuilder    requestBuilder
	createFormBuilder func(io.Writer) formBuilder

	mu sync.Mutex
}

// NewClient creates new OpenAI API client.
func NewClient(authToken string) *Client {
	config := DefaultConfig(authToken)
	return NewClientWithConfig(config)
}

// NewClientWithConfig creates new OpenAI API client for specified config.
func NewClientWithConfig(config ClientConfig) *Client {
	return &Client{
		config:         config,
		requestBuilder: newRequestBuilder(),
		createFormBuilder: func(body io.Writer) formBuilder {
			return newFormBuilder(body)
		},
		mu: sync.Mutex{},
	}
}

// NewOrgClient creates new OpenAI API client for specified Organization ID.
//
// Deprecated: Please use NewClientWithConfig.
func NewOrgClient(authToken, org string) *Client {
	config := DefaultConfig(authToken)
	config.OrgID = org
	return NewClientWithConfig(config)
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")
	// Azure API Key authentication
	if c.config.APIType == APITypeAzure {
		req.Header.Set(AzureAPIKeyHeader, c.config.authToken)
	} else {
		// OpenAI or Azure AD authentication
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))
	}

	// Check whether Content-Type is already set, Upload Files API requires
	// Content-Type == multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	if len(c.config.OrgID) > 0 {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}

	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil || errRes.Error == nil {
			reqErr := RequestError{
				StatusCode: res.StatusCode,
				Err:        err,
			}
			return fmt.Errorf("error, %w", &reqErr)
		}
		errRes.Error.StatusCode = res.StatusCode
		return fmt.Errorf("error, status code: %d, message: %w", res.StatusCode, errRes.Error)
	}

	if v != nil {
		if err = json.NewDecoder(res.Body).Decode(v); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) fullURL(suffix string) string {
	// /openai/deployments/{engine}/chat/completions?api-version={api_version}
	if c.config.APIType == APITypeAzure || c.config.APIType == APITypeAzureAD {
		baseURL := c.config.BaseURL
		baseURL = strings.TrimRight(baseURL, "/")
		return fmt.Sprintf("%s/%s/%s/%s%s?api-version=%s",
			baseURL, azureAPIPrefix, azureDeploymentsPrefix, c.config.Engine, suffix, c.config.APIVersion)
	}

	// c.config.APIType == APITypeOpenAI || c.config.APIType == ""
	return fmt.Sprintf("%s%s", c.config.BaseURL, suffix)
}

func (c *Client) newStreamRequest(
	ctx context.Context,
	method string,
	urlSuffix string,
	body any) (*http.Request, error) {
	req, err := c.requestBuilder.build(ctx, method, c.fullURL(urlSuffix), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/reference#authentication
	// Azure API Key authentication
	if c.config.APIType == APITypeAzure {
		req.Header.Set(AzureAPIKeyHeader, c.config.authToken)
	} else {
		// OpenAI or Azure AD authentication
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))
	}
	return req, nil
}

// SetConfig update client config.
func (c *Client) SetConfig(config ClientConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = config
}

// SetAuthToken update authToken.
func (c *Client) SetAuthToken(authToken string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.authToken = authToken
}

// SetOrgID update orgID.
func (c *Client) SetOrgID(orgID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.OrgID = orgID
}

// SetHTTPClient update http client.
func (c *Client) SetHTTPClient(client *http.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.HTTPClient = client
}
