package openai_test

import (
	"encoding/json"
	"testing"

	"github.com/meguminnnnnnnnn/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestUsageUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"prompt_tokens": 10,
		"completion_tokens": 20,
		"total_tokens": 30,
		"prompt_tokens_details": {
			"cached_tokens": 15
		},
		"completion_tokens_details": {
			"audio_tokens": 10
		},
		"extra_field": "extra_value"
	}`)

	usage := &openai.Usage{}
	err := json.Unmarshal(data, usage)
	assert.NoError(t, err)
	assert.Equal(t, 10, usage.PromptTokens)
	assert.Equal(t, 20, usage.CompletionTokens)
	assert.Equal(t, 30, usage.TotalTokens)
	assert.NotNil(t, usage.PromptTokensDetails)
	assert.Equal(t, 15, usage.PromptTokensDetails.CachedTokens)
	assert.NotNil(t, usage.CompletionTokensDetails)
	assert.Equal(t, 10, usage.CompletionTokensDetails.AudioTokens)
	assert.Len(t, usage.ExtraFields, 1)

	var extraValue string
	err = json.Unmarshal(usage.ExtraFields["extra_field"], &extraValue)
	assert.NoError(t, err)
	assert.Equal(t, "extra_value", extraValue)
}
