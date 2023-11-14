package openai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const billingUsageSuffix = "/billing/usage"

type CostLineItemResponse struct {
	Name string  `json:"name"`
	Cost float64 `json:"cost"` // in cents
}

type DailyCostResponse struct {
	TimestampRaw float64                `json:"timestamp"`
	LineItems    []CostLineItemResponse `json:"line_items"`

	Time time.Time `json:"-"`
}

type BillingUsageResponse struct {
	Object     string              `json:"object"`
	DailyCosts []DailyCostResponse `json:"daily_costs"`
	TotalUsage float64             `json:"total_usage"` // in cents

	httpHeader
}

// currently the OpenAI usage API is not publicly documented and will explictly
// reject requests using an API key authorization. however, it can be utilized
// logging into https://platform.openai.com/usage and retrieving your session
// key from the browser console. session keys have the form 'sess-<keytext>'.
var (
	BillingAPIKeyNotAllowedErrMsg = "Your request to GET /dashboard/billing/usage must be made with a session key (that is, it can only be made from the browser)." //nolint:lll
	ErrSessKeyRequired            = errors.New("an OpenAI API key cannot be used for this request; a session key is required instead")                              //nolint:lll
)

// GetBillingUsage — API call to Get billing usage details.
func (c *Client) GetBillingUsage(ctx context.Context, startDate time.Time,
	endDate time.Time) (response BillingUsageResponse, err error) {
	startDateArg := fmt.Sprintf("start_date=%v", startDate.Format(time.DateOnly))
	endDateArg := fmt.Sprintf("end_date=%v", endDate.Format(time.DateOnly))
	queryParams := fmt.Sprintf("%v&%v", startDateArg, endDateArg)
	urlSuffix := fmt.Sprintf("%v?%v", billingUsageSuffix, queryParams)

	req, err := c.newRequest(ctx, http.MethodGet, c.fullDashboardURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	if err != nil {
		if strings.Contains(err.Error(), BillingAPIKeyNotAllowedErrMsg) {
			err = ErrSessKeyRequired
		}
		return
	}

	for idx, d := range response.DailyCosts {
		dTime := time.Unix(int64(d.TimestampRaw), 0)
		response.DailyCosts[idx].Time = dTime
	}

	return
}
