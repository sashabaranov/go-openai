package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

const (
	TestTotCost    = float64(126.234)
	TestEndDate    = "2023-11-30"
	TestStartDate  = "2023-11-01"
	TestSessionKey = "sess-whatever"
	TestAPIKey     = "sk-whatever"
)

func TestBillingUsageAPIKey(t *testing.T) {
	client, server, teardown := setupOpenAITestServerWithAuth(TestAPIKey)
	defer teardown()
	server.RegisterHandler("/dashboard/billing/usage", handleBillingEndpoint)

	ctx := context.Background()

	endDate, err := time.Parse(openai.DateOnly, TestEndDate)
	checks.NoError(t, err)
	startDate, err := time.Parse(openai.DateOnly, TestStartDate)
	checks.NoError(t, err)

	_, err = client.GetBillingUsage(ctx, startDate, endDate)
	checks.HasError(t, err)
}

func TestBillingUsageSessKey(t *testing.T) {
	client, server, teardown := setupOpenAITestServerWithAuth(TestSessionKey)
	defer teardown()
	server.RegisterHandler("/dashboard/billing/usage", handleBillingEndpoint)

	ctx := context.Background()
	endDate, err := time.Parse(openai.DateOnly, TestEndDate)
	checks.NoError(t, err)
	startDate, err := time.Parse(openai.DateOnly, TestStartDate)
	checks.NoError(t, err)

	resp, err := client.GetBillingUsage(ctx, startDate, endDate)
	checks.NoError(t, err)

	if resp.TotalUsage != TestTotCost {
		t.Errorf("expected total cost %v but got %v", TestTotCost,
			resp.TotalUsage)
	}
	for idx, dc := range resp.DailyCosts {
		if dc.Time.Compare(startDate) < 0 {
			t.Errorf("expected daily cost%v date(%v) before start date %v", idx,
				dc.Time, TestStartDate)
		}
		if dc.Time.Compare(endDate) > 0 {
			t.Errorf("expected daily cost%v date(%v) after end date %v", idx,
				dc.Time, TestEndDate)
		}
	}
}

// handleBillingEndpoint Handles the billing usage endpoint by the test server.
func handleBillingEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if strings.Contains(r.Header.Get("Authorization"), TestAPIKey) {
		http.Error(w, openai.BillingAPIKeyNotAllowedErrMsg, http.StatusUnauthorized)
		return
	}

	var resBytes []byte

	dailyCosts := make([]openai.DailyCostResponse, 0)

	d, _ := time.Parse(openai.DateOnly, TestStartDate)
	d = d.Add(24 * time.Hour)
	dailyCosts = append(dailyCosts, openai.DailyCostResponse{
		TimestampRaw: float64(d.Unix()),
		LineItems: []openai.CostLineItemResponse{
			{Name: "GPT-4 Turbo", Cost: 0.12},
			{Name: "Audio models", Cost: 0.24},
		},
		Time: time.Time{},
	})
	d = d.Add(24 * time.Hour)
	dailyCosts = append(dailyCosts, openai.DailyCostResponse{
		TimestampRaw: float64(d.Unix()),
		LineItems: []openai.CostLineItemResponse{
			{Name: "image models", Cost: 0.56},
		},
		Time: time.Time{},
	})
	res := &openai.BillingUsageResponse{
		Object:     "list",
		DailyCosts: dailyCosts,
		TotalUsage: TestTotCost,
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(resBytes))
}
