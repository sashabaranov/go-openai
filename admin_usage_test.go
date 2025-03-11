package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestAdminUsageCost(t *testing.T) {
	costObject := "page"
	bucketObject := "bucket"
	startTime := int64(1711471533)
	endTime := int64(1711471534)
	resultObject := "organization.costs.result"
	amountValue := 50.23
	amountCurrency := "usd"
	lineItem := "Image Models"
	projectID := "project_abc"
	organizationID := "organization_id"

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/organization/costs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.AdminUsageCostResult{
					Object: resultObject,
					Data: []openai.AdminUsageCostBucket{
						{
							Object:    bucketObject,
							StartTime: startTime,
							EndTime:   endTime,
							Results: []openai.AdminUsageCost{
								{
									Object: costObject,
									Amount: openai.AdminUsageCostAmount{
										Value:    amountValue,
										Currency: amountCurrency,
									},
									LineItem:       &lineItem,
									ProjectID:      &projectID,
									OrganizationID: &organizationID,
								},
							},
						},
					},
					HasMore:  false,
					NextPage: nil,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	ctx := context.Background()

	t.Run("GetAdminUsageCost", func(t *testing.T) {
		request := openai.AdminUsageCostRequest{
			StartTime: startTime,
		}
		costResult, err := client.GetAdminUsageCost(ctx, request)
		checks.NoError(t, err)

		if costResult.Object != resultObject {
			t.Errorf("unexpected object: %v", costResult.Object)
		}

		if len(costResult.Data) != 1 {
			t.Errorf("unexpected data length: %v", len(costResult.Data))
		}

		bucket := costResult.Data[0]
		if bucket.Object != bucketObject {
			t.Errorf("unexpected bucket object: %v", bucket.Object)
		}

		if bucket.StartTime != startTime {
			t.Errorf("unexpected start time: %v", bucket.StartTime)
		}

		if bucket.EndTime != endTime {
			t.Errorf("unexpected end time: %v", bucket.EndTime)
		}

		if len(bucket.Results) != 1 {
			t.Errorf("unexpected results length: %v", len(bucket.Results))
		}

		cost := bucket.Results[0]
		if cost.Object != costObject {
			t.Errorf("unexpected cost object: %v", cost.Object)
		}

		if cost.Amount.Value != amountValue {
			t.Errorf("unexpected amount value: %v", cost.Amount.Value)
		}

		if cost.Amount.Currency != amountCurrency {
			t.Errorf("unexpected amount currency: %v", cost.Amount.Currency)
		}
	})
}
