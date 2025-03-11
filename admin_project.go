package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	adminProjectSuffix = "/organization/projects"
)

// AdminProject represents an Admin Project object.
type AdminProject struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Name       string `json:"name"`
	CreatedAt  int64  `json:"created_at"`
	ArchivedAt *int64 `json:"archived_at"`
	Status     string `json:"status"`

	httpHeader
}

// AdminProjectList represents a list of Admin Projects.
type AdminProjectList struct {
	Object  string         `json:"object"`
	Data    []AdminProject `json:"data"`
	FirstID string         `json:"first_id"`
	LastID  string         `json:"last_id"`
	HasMore bool           `json:"has_more"`

	httpHeader
}

// ListAdminProjects lists Admin Projects associated with the organization.
func (c *Client) ListAdminProjects(
	ctx context.Context,
	limit *int,
	after *string,
	includeArchived *bool,
) (response AdminProjectList, err error) {
	urlValues := url.Values{}
	if limit != nil {
		urlValues.Set("limit", fmt.Sprintf("%d", *limit))
	}
	if after != nil {
		urlValues.Set("after", *after)
	}
	if includeArchived != nil {
		urlValues.Set("include_archived", fmt.Sprintf("%t", *includeArchived))
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := adminProjectSuffix + encodedValues
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CreateAdminProject creates an Admin Project associated with the organization.
func (c *Client) CreateAdminProject(
	ctx context.Context,
	name string,
) (response AdminProject, err error) {
	request := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}

	urlSuffix := adminProjectSuffix
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveAdminProject retrieves an Admin Project associated with the organization.
func (c *Client) RetrieveAdminProject(
	ctx context.Context,
	id string,
) (response AdminProject, err error) {
	urlSuffix := adminProjectSuffix + "/" + id
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ModifyAdminProject modifies an Admin Project associated with the organization.
func (c *Client) ModifyAdminProject(
	ctx context.Context,
	id string,
	name string,
) (response AdminProject, err error) {
	request := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}

	urlSuffix := adminProjectSuffix + "/" + id
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// UpdateAdminProject updates an Admin Project associated with the organization.
func (c *Client) ArchiveAdminProject(
	ctx context.Context,
	id string,
) (response AdminProject, err error) {
	urlSuffix := adminProjectSuffix + "/" + id + "/archive"
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
