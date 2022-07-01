package gogpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// CompletionRequest represents a request structure for completion API
type CompletionRequest struct {
	Model            *string        `json:"model,omitempty"`
	Prompt           string         `json:"prompt,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	Temperature      float32        `json:"temperature,omitempty"`
	TopP             float32        `json:"top_p,omitempty"`
	N                int            `json:"n,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
	LogProbs         int            `json:"logprobs,omitempty"`
	Echo             bool           `json:"echo,omitempty"`
	Stop             []string       `json:"stop,omitempty"`
	PresencePenalty  float32        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32        `json:"frequency_penalty,omitempty"`
	BestOf           int            `json:"best_of,omitempty"`
	LogitBias        map[string]int `json:"logit_bias,omitempty"`
	User             string         `json:"user,omitempty"`
}

// Choice represents one of possible completions
type Choice struct {
	Text         string        `json:"text"`
	Index        int           `json:"index"`
	FinishReason string        `json:"finish_reason"`
	LogProbs     LogprobResult `json:"logprobs"`
}

// LogprobResult represents logprob result of Choice
type LogprobResult struct {
	Tokens        []string             `json:"tokens"`
	TokenLogprobs []float32            `json:"token_logprobs"`
	TopLogprobs   []map[string]float32 `json:"top_logprobs"`
	TextOffset    []int                `json:"text_offset"`
}

// CompletionUsage represents Usage of CompletionResponse
type CompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CompletionResponse represents a response structure for completion API
type CompletionResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created uint64          `json:"created"`
	Model   string          `json:"model"`
	Choices []Choice        `json:"choices"`
	Usage   CompletionUsage `json:"usage"`
}

// CreateCompletion — API call to create a completion. This is the main endpoint of the API. Returns new text as well as, if requested, the probabilities over each alternative token at each position.
func (c *Client) CreateCompletion(ctx context.Context, engineID string, request CompletionRequest) (response CompletionResponse, err error) {
	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	urlSuffix := fmt.Sprintf("/engines/%s/completions", engineID)
	req, err := http.NewRequest("POST", c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &response)
	return
}

// CreateCompletionWithFineTunedModel - API call to create a completion with a fine tuned model
// See https://beta.openai.com/docs/guides/fine-tuning/use-a-fine-tuned-model
// In this case, the model is specified in the CompletionRequest object.
func (c *Client) CreateCompletionWithFineTunedModel(ctx context.Context, request CompletionRequest) (response CompletionResponse, err error) {
	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	urlSuffix := "/completions"
	req, err := http.NewRequest("POST", c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &response)
	return
}
