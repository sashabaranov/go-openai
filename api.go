package gogpt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiURLv1 = "https://api.openai.com/v1"

func newTransport() *http.Client {
	return &http.Client{
		Timeout: time.Minute,
	}
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

// NewOrgClient creates new OpenAI API client for specified Organization ID.
func NewOrgClient(authToken, org string) *Client {
	return &Client{
		BaseURL:    apiURLv1,
		HTTPClient: newTransport(),
		authToken:  authToken,
		idOrg:      org,
	}
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
			return fmt.Errorf("error, status code: %d", res.StatusCode)
		}
		return fmt.Errorf("error, status code: %d, message: %s", res.StatusCode, errRes.Error.Message)
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
