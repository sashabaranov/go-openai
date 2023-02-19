package gogpt

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const apiURLv1 = "https://api.openai.com/v1"

func newTransport() *http.Client {
	return &http.Client{}
}

// Client is OpenAI GPT-3 API client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	authToken  string
	idOrg      string
}

// NewClient creates new OpenAI API client.
func NewClient(authToken string) *Client {
	return &Client{
		BaseURL:    apiURLv1,
		HTTPClient: newTransport(),
		authToken:  authToken,
		idOrg:      "",
	}
}

// NewClientWithTransport creates new OpenAI API client with the provided http.Client.
func NewClientWithTransport(authToken string, transport *http.Client) *Client {
	return &Client{
		BaseURL:    apiURLv1,
		HTTPClient: transport,
		authToken:  authToken,
		idOrg:      "",
	}
}

// NewOrgClient creates new OpenAI API client for specified Organization ID.
func NewOrgClient(authToken, org string) *Client {
	return &Client{
		BaseURL:    apiURLv1,
		HTTPClient: newTransport(),
		authToken:  authToken,
		idOrg:      org,
	}
}

// NewOrgClientWithTransport creates new OpenAI API client for specified Organization ID with the provided http.Client.
func NewOrgClientWithTransport(authToken, org string, transport *http.Client) *Client {
	return &Client{
		BaseURL:    apiURLv1,
		HTTPClient: transport,
		authToken:  authToken,
		idOrg:      org,
	}
}

// WithTransport returns the OpenAI API client configured to use the provided http.Client.
func (c *Client) WithTransport(transport *http.Client) *Client {
	c.HTTPClient = transport
	return c
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))

	// Check whether Content-Type is already set, Upload Files API requires
	// Content-Type == multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	if len(c.idOrg) > 0 {
		req.Header.Set("OpenAI-Organization", c.idOrg)
	}

	res, err := c.HTTPClient.Do(req)
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
		if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) fullURL(suffix string) string {
	return fmt.Sprintf("%s%s", c.BaseURL, suffix)
}
