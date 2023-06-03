package openai_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

type memRateLimiterWaitTestcase struct {
	name                  string
	apiType               APIType
	model                 string
	totalRequests         int
	concurrency           int
	tokensPerRequest      int
	checkCostFunc         func(cost time.Duration) bool
	customRequestLimiters map[string]*rate.Limiter
	customTokensLimiters  map[string]*rate.Limiter
	wantErr               error
}

func newMemRateLimiter(testcase memRateLimiterWaitTestcase) *MemRateLimiter {
	r := NewMemRateLimiter(testcase.apiType)
	if testcase.customRequestLimiters != nil {
		for key, val := range testcase.customRequestLimiters {
			r.RequestLimiters[key] = val
		}
	}

	if testcase.customTokensLimiters != nil {
		for key, val := range testcase.customTokensLimiters {
			r.TokensLimiters[key] = val
		}
	}

	return r
}

func runMemRateLimiterWaitTestCase(tt *testing.T, testcase memRateLimiterWaitTestcase) {
	r := newMemRateLimiter(testcase)
	wg := sync.WaitGroup{}
	wg.Add(testcase.totalRequests)
	for j := 0; j < testcase.concurrency; j++ {
		go func() {
			defer wg.Done()
			err := r.Wait(context.Background(), testcase.model, testcase.tokensPerRequest)
			if err != nil && testcase.wantErr == nil {
				tt.Errorf("Wait() error = %v, want nil", err)
				return
			}
			if err != nil && testcase.wantErr != nil && err.Error() != testcase.wantErr.Error() {
				tt.Errorf("Wait() error = %v, want %v", err, testcase.wantErr)
				return
			}
		}()
	}
	wg.Wait()
}

func TestMemRateLimiter_Wait(t *testing.T) {
	testcases := []memRateLimiterWaitTestcase{
		{
			name:             "test under request limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    100,
			concurrency:      100,
			tokensPerRequest: 0,
		},
		{
			name:             "test equal request limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    300,
			concurrency:      300,
			tokensPerRequest: 0,
		},
		{
			name:             "test over request limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    310,
			concurrency:      310,
			tokensPerRequest: 0,
			checkCostFunc: func(cost time.Duration) bool {
				return cost.Seconds() >= 1
			},
		},
		{
			name:             "test unknown model request limit",
			apiType:          APITypeAzure,
			model:            "unknown",
			totalRequests:    300,
			concurrency:      300,
			tokensPerRequest: 0,
		},
		{
			name:             "test unknown model tokens limit",
			apiType:          APITypeAzure,
			model:            "unknown",
			totalRequests:    10,
			concurrency:      10,
			tokensPerRequest: 12200,
			checkCostFunc: func(cost time.Duration) bool {
				return cost.Seconds() >= 1
			},
		},
		{
			name:             "test under tokens limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    10,
			concurrency:      10,
			tokensPerRequest: 10000,
		},
		{
			name:             "test equal tokens limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    10,
			concurrency:      10,
			tokensPerRequest: 12000,
		},
		{
			name:             "test over tokens limit",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    10,
			concurrency:      10,
			tokensPerRequest: 12800,
			checkCostFunc: func(cost time.Duration) bool {
				return cost.Seconds() >= 1
			},
		},
		{
			name:             "test massive tokens",
			apiType:          APITypeAzure,
			model:            GPT3Dot5Turbo,
			totalRequests:    1,
			concurrency:      1,
			tokensPerRequest: 12200000,
			wantErr:          errors.New("rate: Wait(n=12200000) exceeds limiter's burst 120000"),
		},
		{
			name:          "test unlimited model request limit",
			apiType:       APITypeAzure,
			model:         "unlimited",
			totalRequests: 1,
			concurrency:   1,
			customRequestLimiters: map[string]*rate.Limiter{
				"unlimited": nil,
			},
		},
		{
			name:             "test unlimited model tokens limit",
			apiType:          APITypeAzure,
			model:            "unlimited",
			totalRequests:    1,
			concurrency:      1,
			tokensPerRequest: 12200000,
			customTokensLimiters: map[string]*rate.Limiter{
				"unlimited": nil,
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(tt *testing.T) {
			start := time.Now()
			runMemRateLimiterWaitTestCase(tt, testcase)
			elapsed := time.Since(start)
			if testcase.checkCostFunc != nil && !testcase.checkCostFunc(elapsed) {
				tt.Errorf("Wait() cost time = %v, checkCostFunc %v", elapsed, testcase.checkCostFunc(elapsed))
			}
		})
	}
}

// TestRateLimitedChatCompletions test the rate limiter works with chat completions.
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
	wg.Add(3)
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err = client.CreateChatCompletion(ctx, req)
		checks.NoError(t, err, "CreateChatCompletion error")
		wg.Done()
	}
	wg.Wait()
	elapsed := int(time.Since(start) / time.Second)
	if elapsed != 0 {
		t.Errorf("Wait() cost time = %v, want %v", elapsed, 0)
	}
}

func TestWaitForRequest(t *testing.T) {
	testcases := []struct {
		name    string
		ctx     context.Context
		request TokenCountable
		model   string
		wantErr error
	}{
		{
			name:    "test request is nil",
			model:   "unknown",
			ctx:     context.Background(),
			wantErr: errors.New("request is nil"),
		},
		{
			name:    "test request is nil",
			model:   "unknown",
			request: EmbeddingRequest{},
			wantErr: errors.New("context is nil"),
		},
		{
			name:    "test request is nil",
			model:   "unknown",
			ctx:     context.Background(),
			wantErr: errors.New("request is nil"),
		},
		{
			name:  "test1",
			model: "unknown",
			ctx:   context.Background(),
			request: EmbeddingRequest{
				Input: []string{
					"The food was delicious and the waiter",
					"Other examples of embedding request",
				},
				Model: AdaEmbeddingV2,
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(tt *testing.T) {
			r := NewMemRateLimiter(APITypeAzure)
			err := r.WaitForRequest(testcase.ctx, testcase.model, testcase.request)
			if err != nil && testcase.wantErr == nil {
				tt.Fatalf("Tokens() returned unexpected error: %v", err)
			}

			if err != nil && testcase.wantErr != nil && err.Error() != testcase.wantErr.Error() {
				tt.Fatalf("Tokens() returned unexpected error: %v, want: %v", err, testcase.wantErr)
			}
		})
	}
}

func TestWaitForRequestConcurrency(t *testing.T) {
	testcases := []struct {
		name    string
		ctx     context.Context
		request TokenCountable
		model   string
		wantErr error
	}{
		{
			name:    "test request is nil",
			model:   "unknown",
			ctx:     context.Background(),
			wantErr: errors.New("request is nil"),
		},
		{
			name:  "test request is nil",
			model: "unknown",
			ctx:   context.Background(),
			request: EmbeddingRequest{
				Input: []string{
					"The food was delicious and the waiter",
					"Other examples of embedding request",
				},
				Model: AdaEmbeddingV2,
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(tt *testing.T) {
			r := NewMemRateLimiter(APITypeAzure)
			err := r.WaitForRequest(testcase.ctx, testcase.model, testcase.request)
			if err != nil && testcase.wantErr == nil {
				tt.Fatalf("Tokens() returned unexpected error: %v", err)
			}

			if err != nil && testcase.wantErr != nil && err.Error() != testcase.wantErr.Error() {
				tt.Fatalf("Tokens() returned unexpected error: %v, want: %v", err, testcase.wantErr)
			}
		})
	}
}

func TestNilRequestLimiters(t *testing.T) {
	r := NewMemRateLimiter(APITypeAzure)
	r.RequestLimiters = nil
	err := r.Wait(context.Background(), GPT3Dot5Turbo, 0)
	checks.NoError(t, err, "Wait error")
}

func TestNilTokensLimiters(t *testing.T) {
	r := NewMemRateLimiter(APITypeAzure)
	r.TokensLimiters = nil
	err := r.Wait(context.Background(), GPT3Dot5Turbo, 0)
	checks.NoError(t, err, "Wait error")
}

func TestWaitForRateLimitTokensErr(t *testing.T) {
	r := NewMemRateLimiter(APITypeAzure)
	r.TokensLimiters = nil
	err := r.Wait(context.Background(), GPT3Dot5Turbo, 0)
	checks.NoError(t, err, "Wait error")
}
