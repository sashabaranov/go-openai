package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	utils "github.com/sashabaranov/go-openai/internal"
)

// Client is OpenAI GPT-3 API client.
type Client struct {
	config ClientConfig

	requestBuilder    utils.RequestBuilder
	createFormBuilder func(io.Writer) utils.FormBuilder
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
		requestBuilder: utils.NewRequestBuilder(),
		createFormBuilder: func(body io.Writer) utils.FormBuilder {
			return utils.NewFormBuilder(body)
		},
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

func (c *Client) sendRequest(req *http.Request, v any) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")

	// Check whether Content-Type is already set, Upload Files API requires
	// Content-Type == multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	c.setCommonHeaders(req)

	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if isFailureStatusCode(res) {
		return c.handleErrorResp(res)
	}

	return decodeResponse(res.Body, v)
}

func (c *Client) setCommonHeaders(req *http.Request) {
	// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/reference#authentication
	// Azure API Key authentication
	if c.config.APIType == APITypeAzure {
		req.Header.Set(AzureAPIKeyHeader, c.config.authToken)
	} else {
		// OpenAI or Azure AD authentication
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))
	}
	if c.config.OrgID != "" {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}
}

func isFailureStatusCode(resp *http.Response) bool {
	return resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest
}

func decodeResponse(body io.Reader, v any) error {
	if v == nil {
		return nil
	}

	if result, ok := v.(*string); ok {
		return decodeString(body, result)
	}
	return json.NewDecoder(body).Decode(v)
}

func decodeString(body io.Reader, output *string) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	*output = string(b)
	return nil
}

// fullURL returns full URL for request.
// args[0] is model name, if API type is Azure, model name is required to get deployment name.
func (c *Client) fullURL(suffix string, args ...any) string {
	// /openai/deployments/{model}/chat/completions?api-version={api_version}
	if c.config.APIType == APITypeAzure || c.config.APIType == APITypeAzureAD {
		baseURL := c.config.BaseURL
		baseURL = strings.TrimRight(baseURL, "/")
		// if suffix is /models change to {endpoint}/openai/models?api-version=2022-12-01
		// https://learn.microsoft.com/en-us/rest/api/cognitiveservices/azureopenaistable/models/list?tabs=HTTP
		if strings.Contains(suffix, "/models") {
			return fmt.Sprintf("%s/%s%s?api-version=%s", baseURL, azureAPIPrefix, suffix, c.config.APIVersion)
		}
		azureDeploymentName := "UNKNOWN"
		if len(args) > 0 {
			model, ok := args[0].(string)
			if ok {
				azureDeploymentName = c.config.GetAzureDeploymentByModel(model)
			}
		}
		return fmt.Sprintf("%s/%s/%s/%s%s?api-version=%s",
			baseURL, azureAPIPrefix, azureDeploymentsPrefix,
			azureDeploymentName, suffix, c.config.APIVersion,
		)
	}

	// c.config.APIType == APITypeOpenAI || c.config.APIType == ""
	return fmt.Sprintf("%s%s", c.config.BaseURL, suffix)
}

func (c *Client) newStreamRequest(
	ctx context.Context,
	method string,
	urlSuffix string,
	body any,
	model string) (*http.Request, error) {
	req, err := c.requestBuilder.Build(ctx, method, c.fullURL(urlSuffix, model), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	c.setCommonHeaders(req)
	return req, nil
}

func (c *Client) handleErrorResp(resp *http.Response) error {
	var errRes ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errRes)
	if err != nil || errRes.Error == nil {
		reqErr := &RequestError{
			HTTPStatusCode: resp.StatusCode,
			Err:            err,
		}
		if errRes.Error != nil {
			reqErr.Err = errRes.Error
		}
		return reqErr
	}

	errRes.Error.HTTPStatusCode = resp.StatusCode
	return errRes.Error
}
