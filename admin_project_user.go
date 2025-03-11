package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AdminProjectUser represents a user associated with a project.
type AdminProjectUser struct {
	Object  string `json:"object"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	AddedAt int64  `json:"added_at"`

	httpHeader
}

// AdminProjectUserList represents a list of users associated with a project.
type AdminProjectUserList struct {
	Object  string             `json:"object"`
	Data    []AdminProjectUser `json:"data"`
	FirstID string             `json:"first_id"`
	LastID  string             `json:"last_id"`
	HasMore bool               `json:"has_more"`

	httpHeader
}

// AdminProjectDeleteResponse represents the response when deleting a project.
type AdminProjectDeleteResponse struct {
	Object  string `json:"object"`
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

// ListAdminProjectUsers lists users associated with a project.
func (c *Client) ListAdminProjectUsers(
	ctx context.Context,
	projectID string,
	limit *int,
	after *string,
) (response AdminProjectUserList, err error) {
	urlValues := url.Values{}
	urlValues.Add("project", projectID)
	if limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if after != nil {
		urlValues.Add("after", *after)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("/v1/projects/%s/users%s", projectID, encodedValues)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CreateProjectUser creates a user associated with a project.
func (c *Client) CreateAdminProjectUser(
	ctx context.Context,
	projectID string,
	userID string,
	role string,
) (response AdminProjectUser, err error) {
	request := struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}{
		UserID: userID,
		Role:   role,
	}

	urlSuffix := fmt.Sprintf("/v1/projects/%s/users", projectID)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveProjectUser retrieves a user associated with a project.
func (c *Client) RetrieveAdminProjectUser(
	ctx context.Context,
	projectID string,
	userID string,
) (response AdminProjectUser, err error) {
	urlSuffix := fmt.Sprintf("/v1/projects/%s/users/%s", projectID, userID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ModifyProjectUser modifies a user associated with a project.
func (c *Client) ModifyAdminProjectUser(
	ctx context.Context,
	projectID string,
	userID string,
	role string,
) (response AdminProjectUser, err error) {
	request := struct {
		Role string `json:"role"`
	}{
		Role: role,
	}

	urlSuffix := fmt.Sprintf("/v1/projects/%s/users/%s", projectID, userID)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// DeleteProjectUser deletes a user associated with a project.
func (c *Client) DeleteAdminProjectUser(
	ctx context.Context,
	projectID string,
	userID string,
) (response AdminProjectDeleteResponse, err error) {
	urlSuffix := fmt.Sprintf("/v1/projects/%s/users/%s", projectID, userID)

	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
