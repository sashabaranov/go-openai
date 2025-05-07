package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	adminKeysSuffix = "/organization/admin_api_keys"
)

// AdminKey represents an Admin API Key.
type AdminKey struct {
	Object        string        `json:"object"`
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	RedactedValue string        `json:"redacted_value"`
	CreatedAt     int64         `json:"created_at"`
	Owner         AdminKeyOwner `json:"owner"`
	Value         *string       `json:"value"`

	httpHeader
}

// AdminKeyOwner represents the owner of an Admin API Key.
type AdminKeyOwner struct {
	Type      string `json:"type"`
	Object    string `json:"object"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	Role      string `json:"role"`
}

// AdminKeyList represents a list of Admin API Keys.
type AdminKeyList struct {
	Object    string     `json:"object"`
	AdminKeys []AdminKey `json:"data"`
	FirstID   string     `json:"first_id"`
	LastID    string     `json:"last_id"`
	HasMore   bool       `json:"has_more"`
	httpHeader
}

// AdminKeyDeleteResponse represents the response from deleting an Admin API Key.
type AdminKeyDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

// ListAdminKeys lists Admin API Keys associated with the organization.
func (c *Client) ListAdminKeys(
	ctx context.Context,
	limit *int,
	order *string,
	after *string,
) (response AdminKeyList, err error) {
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

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := adminKeysSuffix + encodedValues
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CreateAdminKey creates a new Admin API Key.
func (c *Client) CreateAdminKey(
	ctx context.Context,
	keyName string,
) (response AdminKey, err error) {
	type KeyName struct {
		Name string `json:"name"`
	}
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(adminKeysSuffix),
		withBody(KeyName{Name: keyName}))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)

	return
}

// RetrieveAdminKey retrieves an Admin API Key.
func (c *Client) RetrieveAdminKey(
	ctx context.Context,
	keyID string,
) (response AdminKey, err error) {
	urlSuffix := fmt.Sprintf("%s/%s", adminKeysSuffix, keyID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// DeleteAdminKey deletes an Admin API Key.
func (c *Client) DeleteAdminKey(
	ctx context.Context,
	keyID string,
) (
	response AdminKeyDeleteResponse, err error,
) {
	urlSuffix := fmt.Sprintf("%s/%s", adminKeysSuffix, keyID)
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
