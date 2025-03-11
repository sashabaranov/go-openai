package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	adminUsersSuffix = "/organization/users"
)

// AdminUser represents a User.
type AdminUser struct {
	Object  string `json:"object"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	AddedAt int64  `json:"added_at"`

	httpHeader
}

// AdminUserList represents a list of Users.
type AdminUserList struct {
	Object  string      `json:"object"`
	User    []AdminUser `json:"data"`
	FirstID string      `json:"first_id"`
	LastID  string      `json:"last_id"`
	HasMore bool        `json:"has_more"`

	httpHeader
}

// AdminUserDeleteResponse represents the response from deleting an User.
type AdminUserDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

// AdminListUsers lists Users associated with the organization.
func (c *Client) ListAdminUsers(
	ctx context.Context,
	limit *int,
	after *string,
	emails *[]string,
) (response AdminUserList, err error) {
	urlValues := url.Values{}
	if limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if after != nil {
		urlValues.Add("after", *after)
	}
	if emails != nil {
		for _, email := range *emails {
			urlValues.Add("emails[]", email)
		}
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := adminUsersSuffix + encodedValues
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ModifyAdminUser modifies an User.
func (c *Client) ModifyAdminUser(
	ctx context.Context,
	id string,
	role string,
) (response AdminUser, err error) {
	type ModifyAdminUserRequest struct {
		Role string `json:"role"`
	}
	urlSuffix := adminUsersSuffix + "/" + id
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix),
		withBody(ModifyAdminUserRequest{Role: role}))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveAdminUser retrieves an User.
func (c *Client) RetrieveAdminUser(
	ctx context.Context,
	id string,
) (response AdminUser, err error) {
	urlSuffix := adminUsersSuffix + "/" + id
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// AdminDeleteUser deletes an User.
func (c *Client) DeleteAdminUser(
	ctx context.Context,
	id string,
) (response AdminUserDeleteResponse, err error) {
	urlSuffix := adminUsersSuffix + "/" + id
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
