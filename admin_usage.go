package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	adminUsageCostSuffix = "/organization/costs"
)

// AdminUsageCostRequest represents a request to get usage costs.
type AdminUsageCostRequest struct {
	StartTime   int64    `json:"start_time"`
	EndTime     *int64   `json:"end_time,omitempty"`
	BucketWidth *string  `json:"bucket_width,omitempty"`
	ProjectIDs  []string `json:"project_ids,omitempty"`
	GroupBy     *string  `json:"group_by,omitempty"`
	Limit       *int     `json:"limit,omitempty"`
	Page        *string  `json:"page,omitempty"`
}

// AdminUsageCost represents a usage cost.
type AdminUsageCost struct {
	Object         string               `json:"object"`
	Amount         AdminUsageCostAmount `json:"amount"`
	LineItem       *string              `json:"line_item"`
	ProjectID      *string              `json:"project_id"`
	OrganizationID *string              `json:"organization_id"`
}

// AdminUsageCostAmount represents the amount of a usage cost.
type AdminUsageCostAmount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// AdminUsageCostBucket represents a bucket of usage costs based on a time range.
type AdminUsageCostBucket struct {
	Object    string           `json:"object"`
	StartTime int64            `json:"start_time"`
	EndTime   int64            `json:"end_time"`
	Results   []AdminUsageCost `json:"results"`
}

// AdminUsageCostResult represents the response from getting usage costs.
type AdminUsageCostResult struct {
	Object   string                 `json:"object"`
	Data     []AdminUsageCostBucket `json:"data"`
	HasMore  bool                   `json:"has_more"`
	NextPage *string                `json:"next_page"`

	httpHeader
}

// GetAdminUsageCost gets usage costs for the organization.
func (c *Client) GetAdminUsageCost(
	ctx context.Context,
	request AdminUsageCostRequest,
) (response AdminUsageCostResult, err error) {
	urlValues := url.Values{}
	urlValues.Add("start_time", fmt.Sprintf("%d", request.StartTime))
	if request.EndTime != nil {
		urlValues.Add("end_time", fmt.Sprintf("%d", *request.EndTime))
	}
	if request.BucketWidth != nil {
		urlValues.Add("bucket_width", *request.BucketWidth)
	}
	if len(request.ProjectIDs) > 0 {
		for _, projectID := range request.ProjectIDs {
			urlValues.Add("project_ids[]", projectID)
		}
	}
	if request.GroupBy != nil {
		urlValues.Add("group_by", *request.GroupBy)
	}
	if request.Limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *request.Limit))
	}
	if request.Page != nil {
		urlValues.Add("page", *request.Page)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := adminUsageCostSuffix + encodedValues
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
