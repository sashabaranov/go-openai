package openai

import (
	"encoding/json"
	"fmt"
	"reflect"

	openai "github.com/meguminnnnnnnnn/go-openai/internal"
)

// common.go defines common types used throughout the OpenAI API.

// Usage Represents the total token usage per request to OpenAI.
type Usage struct {
	PromptTokens            int                        `json:"prompt_tokens"`
	CompletionTokens        int                        `json:"completion_tokens"`
	TotalTokens             int                        `json:"total_tokens"`
	PromptTokensDetails     *PromptTokensDetails       `json:"prompt_tokens_details"`
	CompletionTokensDetails *CompletionTokensDetails   `json:"completion_tokens_details"`
	ExtraFields             map[string]json.RawMessage `json:"-"`
}

func (u *Usage) UnmarshalJSON(data []byte) error {
	if u == nil {
		return fmt.Errorf("usage is nil")
	}

	type Alias Usage
	alias := &Alias{}
	err := json.Unmarshal(data, alias)
	if err != nil {
		return err
	}

	*u = Usage(*alias)

	extra, err := openai.UnmarshalExtraFields(reflect.TypeOf(u), data)
	if err != nil {
		return err
	}

	u.ExtraFields = extra

	return nil
}

// CompletionTokensDetails Breakdown of tokens used in a completion.
type CompletionTokensDetails struct {
	AudioTokens              int `json:"audio_tokens"`
	ReasoningTokens          int `json:"reasoning_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

// PromptTokensDetails Breakdown of tokens used in the prompt.
type PromptTokensDetails struct {
	AudioTokens  int `json:"audio_tokens"`
	CachedTokens int `json:"cached_tokens"`
}
