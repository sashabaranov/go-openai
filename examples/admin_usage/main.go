package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
)

func main() {
	ctx := context.Background()

	// Admin API Keys are different than regular API keys and require a different
	// endpoint to be used. You can find your Admin API key in the OpenAI dashboard.
	config := openai.DefaultConfig(os.Getenv("OPENAI_ADMIN_API_KEY"))
	client := openai.NewClientWithConfig(config)

	// Specify the date range of the usage data you want to retrieve, the end date is optional,
	// but when specified, it should include through the end of the day you want to retrieve.
	startTime := convertDateStringToTimestamp("2025-02-01")
	endTime := convertDateStringToTimestamp("2025-03-01")

	// In this example each bucket represents a day of usage data. To avoid
	// making several requests to get the data for each day, we'll increase
	// the limit to 31 to get all the data in one request.
	limit := 31

	// Create the request object, only StartTime is required.
	req := openai.AdminUsageCostRequest{
		StartTime: startTime,
		EndTime:   &endTime,
		Limit:     &limit,
	}

	// Request the usage data.
	res, err := client.GetAdminUsageCost(ctx, req)
	if err != nil {
		fmt.Printf("error getting openai usage data: %v\n", err)
		return
	}

	// Calculate the total cost of the usage data.
	totalCost := 0.0
	for _, bucket := range res.Data {
		for _, cost := range bucket.Results {
			totalCost += cost.Amount.Value
		}
	}

	fmt.Printf("Total Cost: %f\n", totalCost)
}

// Helper function to convert a date string to a Unix timestamp.
func convertDateStringToTimestamp(date string) int64 {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return t.Unix()
}
