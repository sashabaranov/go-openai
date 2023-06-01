package openai

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Wait(ctx context.Context, model string, tokens int) error
}

// MemRateLimiter is a token bucket based rate limiter for OpenAI API.
type MemRateLimiter struct {
	RequestLimiters     map[string]*rate.Limiter
	DefaultRequestLimit rate.Limit
	TokensLimiters      map[string]*rate.Limiter
	DefaultTokensLimit  rate.Limit
	apiType             APIType
}

func NewMemRateLimiter(apiType APIType) *MemRateLimiter {
	r := &MemRateLimiter{
		apiType: apiType,
	}

	if r.apiType == APITypeAzure || r.apiType == APITypeAzureAD {
		r.DefaultRequestLimit = 300
		r.DefaultTokensLimit = 120000
		r.RequestLimiters = r.newAzureRequestLimiters()
		r.TokensLimiters = r.newAzureTokensLimiters()
	}

	if r.apiType == APITypeOpenAI {
		r.DefaultRequestLimit = 3500
		r.DefaultTokensLimit = 350000
		r.RequestLimiters = r.newOpenAIRequestLimiters()
		r.TokensLimiters = r.newOpenAITokensLimiters()
	}

	return r
}

func (r *MemRateLimiter) Wait(ctx context.Context, model string, tokens int) (err error) {
	err = r.requestWait(ctx, model)
	if err != nil {
		return err
	}

	if tokens == 0 {
		return nil
	}

	return r.tokensWait(ctx, model, tokens)
}

func (r *MemRateLimiter) requestWait(ctx context.Context, model string) (err error) {
	var limiter *rate.Limiter
	var ok bool

	if r.RequestLimiters == nil {
		return nil
	}

	limiter, ok = r.RequestLimiters[model]
	if !ok {
		limiter = rate.NewLimiter(r.DefaultRequestLimit/60, int(r.DefaultRequestLimit))
	}

	// if limiter is nil, it means that the model is not rate limited
	if limiter == nil {
		return nil
	}

	return limiter.WaitN(ctx, 1)
}

func (r *MemRateLimiter) tokensWait(ctx context.Context, model string, tokens int) (err error) {
	var limiter *rate.Limiter
	var ok bool

	if r.TokensLimiters == nil {
		return nil
	}

	limiter, ok = r.TokensLimiters[model]
	if !ok {
		limiter = rate.NewLimiter(r.DefaultTokensLimit/60, int(r.DefaultTokensLimit))
	}

	// if limiter is nil, it means that the model is not rate limited
	if limiter == nil {
		return nil
	}

	return limiter.WaitN(ctx, tokens)
}

// newAzureRequestLimiters creates a map of request limiters for each azure openai model.
// reference: https://learn.microsoft.com/en-us/azure/cognitive-services/openai/quotas-limits#quotas-and-limits-reference
func (r *MemRateLimiter) newAzureRequestLimiters() map[string]*rate.Limiter {
	// The limiters that are not defined here are controlled by the DefaultRequestLimit.
	return map[string]*rate.Limiter{
		GPT3Davinci:       rate.NewLimiter(120/60, 120),
		GPT3Dot5Turbo:     rate.NewLimiter(300/60, 300),
		GPT3Dot5Turbo0301: rate.NewLimiter(300/60, 300),
		GPT4:              rate.NewLimiter(18/60, 18),
		GPT432K:           rate.NewLimiter(18/60, 18),
	}
}

// newAzureTokensLimiters creates a map of tokens limiters for each azure openai model.
func (r *MemRateLimiter) newAzureTokensLimiters() map[string]*rate.Limiter {
	// The limiters that are not defined here are controlled by the DefaultTokensLimit.
	return map[string]*rate.Limiter{
		GPT3Davinci:       rate.NewLimiter(40000/60, 40000),
		GPT3Dot5Turbo:     rate.NewLimiter(120000/60, 120000),
		GPT3Dot5Turbo0301: rate.NewLimiter(120000/60, 120000),
		GPT4:              rate.NewLimiter(10000/60, 10000),
		GPT432K:           rate.NewLimiter(32000/60, 32000),
	}
}

// newOpenAIRequestLimiters creates a map of request limiters for each openai model.
// reference: https://platform.openai.com/docs/guides/rate-limits/overview
func (r *MemRateLimiter) newOpenAIRequestLimiters() map[string]*rate.Limiter {
	// The limiters that are not defined here are controlled by the DefaultRequestLimit.
	return map[string]*rate.Limiter{
		GPT3Davinci:       rate.NewLimiter(3500/60, 3500),
		GPT3Dot5Turbo:     rate.NewLimiter(3500/60, 3500),
		GPT3Dot5Turbo0301: rate.NewLimiter(3500/60, 3500),
		GPT4:              rate.NewLimiter(200/60, 200),
		GPT40314:          rate.NewLimiter(200/60, 200),
		GPT432K:           rate.NewLimiter(20/60, 20),
		GPT432K0314:       rate.NewLimiter(20/60, 20),
	}
}

// newOpenAITokensLimiters creates a map of tokens limiters for each openai model.
func (r *MemRateLimiter) newOpenAITokensLimiters() map[string]*rate.Limiter {
	// The limiters that are not defined here are controlled by the DefaultTokensLimit.
	return map[string]*rate.Limiter{
		GPT3Davinci:            rate.NewLimiter(350000/60, 350000),
		GPT3TextAdaEmbeddingV2: rate.NewLimiter(350000*200/60, 350000*200),
		GPT3Dot5Turbo:          rate.NewLimiter(350000/60, 350000),
		GPT3Dot5Turbo0301:      rate.NewLimiter(350000/60, 350000),
		GPT4:                   rate.NewLimiter(40000/60, 40000),
		GPT432K:                rate.NewLimiter(150000/60, 150000),
	}
}

type TokenCountable interface {
	Tokens() (int, error)
}

func waitForRateLimit(c *Client, ctx context.Context, request TokenCountable, model string) (err error) {
	if c.rateLimiter == nil {
		return nil
	}

	var tokens int
	tokens, err = request.Tokens()
	if err != nil {
		err = fmt.Errorf("failed to get tokens count: %w", err)
		return
	}

	err = c.rateLimiter.Wait(ctx, model, tokens)
	if err != nil {
		err = fmt.Errorf("failed to wait for rate limiter: %w", err)
		return
	}

	return
}
