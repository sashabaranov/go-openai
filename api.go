package gogpt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiURLv1 = "https://api.openai.com/v1"

// Client is OpenAI GPT-3 API client
type Client struct {
	BaseURL    string
	authToken  string
	HTTPClient *http.Client
}

// NewClient creates new OpenAI API client
func NewClient(authToken string) *Client {
	return &Client{
		BaseURL:   apiURLv1,
		authToken: authToken,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error, status code: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}

func (c *Client) fullURL(suffix string) string {
	return fmt.Sprintf("%s%s", c.BaseURL, suffix)
}
