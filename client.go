package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	utils "github.com/sashabaranov/go-openai/internal"
)

var (
	ErrClientEmptyCallbackURL          = errors.New("Error retrieving callback URL (Operation-Location) for image request") //nolint:lll
	ErrClientRetievingCallbackResponse = errors.New("Error retrieving callback response")                                   //nolint:lll
)

// Client is OpenAI GPT-3 API client.
type Client struct {
	config ClientConfig

	requestBuilder    utils.RequestBuilder
	createFormBuilder func(io.Writer) utils.FormBuilder
}

// Azure image request callback response struct.
type CBData []struct {
	URL string `json:"url"`
}
type CBResult struct {
	Data CBData `json:"data"`
}
type CallBackResponse struct {
	Created int64    `json:"created"`
	Expires int64    `json:"expires"`
	ID      string   `json:"id"`
	Result  CBResult `json:"result"`
	Status  string   `json:"status"`
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

	if c.config.APIType == APITypeAzure || c.config.APIType == APITypeAzureAD {
		// Special handling for initial call to Azure DALL-E API.
		if strings.Contains(req.URL.Path, "openai/images/generations") {
			return c.requestImage(res, v)
		}
		// Special handling for callBack to Azure DALL-E API.
		if strings.Contains(req.URL.Path, "openai/operations/images") {
			return c.imageRequestCallback(req, v, res)
		}
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

func (c *Client) requestImage(res *http.Response, v any) error {
	_, err := io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		return err
	}
	callBackURL := res.Header.Get("Operation-Location")
	if callBackURL == "" {
		return ErrClientEmptyCallbackURL
	}
	newReq, err := http.NewRequest("GET", callBackURL, nil)
	if err != nil {
		return err
	}
	return c.sendRequest(newReq, v)
}

// Handle image callback response from Azure DALL-E API.
func (c *Client) imageRequestCallback(req *http.Request, v any, res *http.Response) error {
	// Retry Sleep seconds for Azure DALL-E 2 callback URL.
	var callBackWaitTime = 3

	// Wait for the callBack to complete
	var result *CallBackResponse
	err := json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Status == "" {
		return ErrClientRetievingCallbackResponse
	}
	if result.Status != "Succeeded" {
		time.Sleep(time.Duration(callBackWaitTime) * time.Second)
		req.Header.Add("Retry", "true")
		return c.sendRequest(req, v)
	}

	// Convert the callBack response to the OpenAI ImageResponse
	var urlList []ImageResponseDataInner
	for _, data := range result.Result.Data {
		urlList = append(urlList, ImageResponseDataInner{URL: data.URL})
	}
	converted, err := json.Marshal(ImageResponse{Created: result.Created, Data: urlList})
	if err != nil {
		return err
	}
	return decodeResponse(bytes.NewReader(converted), v)
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
	if c.config.APIType == APITypeAzure || c.config.APIType == APITypeAzureAD {
		baseURL := c.config.BaseURL
		baseURL = strings.TrimRight(baseURL, "/")
		switch {
		case strings.Contains(suffix, "/models"):
			// if suffix is /models change to {endpoint}/openai/models?api-version={api_version}
			// https://learn.microsoft.com/en-us/rest/api/cognitiveservices/azureopenaistable/models/list?tabs=HTTP
			return fmt.Sprintf("%s/%s%s?api-version=%s", baseURL, azureAPIPrefix, suffix, c.config.APIVersion)
		case strings.Contains(suffix, "/images"):
			// if suffix is /images change to {endpoint}openai/images/generations:submit?api-version={api_version}
			return fmt.Sprintf("%s/%s%s:submit?api-version=%s", baseURL, azureAPIPrefix, suffix, c.config.APIVersion)
		default:
			// /openai/deployments/{model}/chat/completions?api-version={api_version}
			azureDeploymentName := "UNKNOWN"
			if len(args) > 0 {
				model, ok := args[0].(string)
				if ok {
					azureDeploymentName = c.config.GetAzureDeploymentByModel(model)
				}
			}
			return fmt.Sprintf("%s/%s/%s/%s%s?api-version=%s", baseURL, azureAPIPrefix, azureDeploymentsPrefix,
				azureDeploymentName, suffix, c.config.APIVersion,
			)
		}
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
