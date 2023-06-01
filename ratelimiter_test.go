package openai_test

import (
	"context"
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
	"sync"
	"testing"
	"time"
)

func TestMemRateLimiter_Wait(t *testing.T) {
	testcases := []struct {
		name             string
		apiType          APIType
		model            string
		totalRequests    int
		concurrency      int
		tokensPerRequest int
		wantCostSeconds  int
	}{
		{
			name:             "test under request limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    100,
			concurrency:      100,
			tokensPerRequest: 0,
			wantCostSeconds:  0,
		},
		{
			name:             "test equal request limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    300,
			concurrency:      300,
			tokensPerRequest: 0,
			wantCostSeconds:  0,
		},
		{
			name:             "test over request limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    305,
			concurrency:      305,
			tokensPerRequest: 0,
			wantCostSeconds:  1,
		},
		{
			name:             "test under tokens limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    10,
			concurrency:      10,
			tokensPerRequest: 10000,
			wantCostSeconds:  0,
		},
		{
			name:             "test equal tokens limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    10,
			concurrency:      10,
			tokensPerRequest: 12000,
			wantCostSeconds:  0,
		},
		{
			name:             "test over tokens limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    10,
			concurrency:      10,
			tokensPerRequest: 12200,
			wantCostSeconds:  1,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(tt *testing.T) {
			r := NewMemRateLimiter(testcase.apiType)
			start := time.Now()
			wg := sync.WaitGroup{}
			wg.Add(testcase.totalRequests)
			for j := 0; j < testcase.concurrency; j++ {
				go func() {
					defer wg.Done()
					err := r.Wait(context.Background(), testcase.model, testcase.tokensPerRequest)
					if err != nil {
						tt.Errorf("Wait() error = %v", err)
						return
					}
				}()
			}
			wg.Wait()
			elapsed := int(time.Since(start) / time.Second)
			if elapsed != testcase.wantCostSeconds {
				tt.Errorf("Wait() cost time = %v, want %v", elapsed, testcase.wantCostSeconds)
			}
		})
	}
}

// TestRateLimitedChatCompletions test the rate limiter works with chat completions
func TestRateLimitedChatCompletions(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/chat/completions", handleChatCompletionEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	config.EnableRateLimiter = true
	client := NewClientWithConfig(config)
	ctx := context.Background()

	req := ChatCompletionRequest{
		MaxTokens: 5,
		Model:     GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	}

	wg := sync.WaitGroup{}
	wg.Add(3675)
	start := time.Now()
	for i := 0; i < 3675; i++ {
		_, err = client.CreateChatCompletion(ctx, req)
		checks.NoError(t, err, "CreateChatCompletion error")
		wg.Done()
	}
	wg.Wait()
	elapsed := int(time.Since(start) / time.Second)
	if elapsed != 3 {
		t.Errorf("Wait() cost time = %v, want %v", elapsed, 3)
	}
}
