package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	adminInvitesSuffix = "/organization/invites"
)

var (
	// adminInviteRoles is a list of valid roles for an Admin Invite.
	adminInviteRoles = []string{"owner", "member"}
)

// AdminInvite represents an Admin Invite.
type AdminInvite struct {
	Object     string               `json:"object"`
	ID         string               `json:"id"`
	Email      string               `json:"email"`
	Role       string               `json:"role"`
	Status     string               `json:"status"`
	InvitedAt  int64                `json:"invited_at"`
	ExpiresAt  int64                `json:"expires_at"`
	AcceptedAt int64                `json:"accepted_at"`
	Projects   []AdminInviteProject `json:"projects"`

	httpHeader
}

// AdminInviteProject represents a project associated with an Admin Invite.
type AdminInviteProject struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// AdminInviteList represents a list of Admin Invites.
type AdminInviteList struct {
	Object       string        `json:"object"`
	AdminInvites []AdminInvite `json:"data"`
	FirstID      string        `json:"first_id"`
	LastID       string        `json:"last_id"`
	HasMore      bool          `json:"has_more"`

	httpHeader
}

// AdminInviteDeleteResponse represents the response from deleting an Admin Invite.
type AdminInviteDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

// ListAdminInvites lists Admin Invites associated with the organization.
func (c *Client) ListAdminInvites(
	ctx context.Context,
	limit *int,
	after *string,
) (response AdminInviteList, err error) {
	urlValues := url.Values{}
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

	urlSuffix := adminInvitesSuffix + encodedValues
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CreateAdminInvite creates a new Admin Invite.
func (c *Client) CreateAdminInvite(
	ctx context.Context,
	email string,
	role string,
	projects *[]AdminInviteProject,
) (response AdminInvite, err error) {
	// Validate the role.
	if !containsSubstr(adminInviteRoles, role) {
		return response, fmt.Errorf("invalid admin role: %s", role)
	}

	// Create the request object.
	request := struct {
		Email    string                `json:"email"`
		Role     string                `json:"role"`
		Projects *[]AdminInviteProject `json:"projects,omitempty"`
	}{
		Email:    email,
		Role:     role,
		Projects: projects,
	}

	urlSuffix := adminInvitesSuffix
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)

	return
}

// RetrieveAdminInvite retrieves an Admin Invite.
func (c *Client) RetrieveAdminInvite(
	ctx context.Context,
	inviteID string,
) (response AdminInvite, err error) {
	urlSuffix := fmt.Sprintf("%s/%s", adminInvitesSuffix, inviteID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// DeleteAdminInvite deletes an Admin Invite.
func (c *Client) DeleteAdminInvite(
	ctx context.Context,
	inviteID string,
) (response AdminInviteDeleteResponse, err error) {
	urlSuffix := fmt.Sprintf("%s/%s", adminInvitesSuffix, inviteID)
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
