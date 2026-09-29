package openai_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestGPT6ChatCompatibility(t *testing.T) { //nolint:gocognit
	models := []struct {
		name string
		none bool
	}{
		{openai.GPT6Dot1Sol, false}, {openai.GPT6Astra, false},
		{openai.GPT6Sol, true}, {openai.GPT6Luna, true},
	}
	for _, model := range models {
		t.Run(model.name, func(t *testing.T) {
			validator := openai.NewReasoningValidator()
			for _, effort := range []string{"", "low", "medium", "high", "xhigh", "max"} {
				if err := validator.Validate(openai.ChatCompletionRequest{Model: model.name, ReasoningEffort: effort}); err != nil {
					t.Errorf("effort %q: %v", effort, err)
				}
			}
			tests := []struct {
				name    string
				request openai.ChatCompletionRequest
				want    error
			}{
				{"minimal", openai.ChatCompletionRequest{ReasoningEffort: "minimal"}, openai.ErrReasoningEffortUnsupported},
				{"invalid effort", openai.ChatCompletionRequest{ReasoningEffort: "ultra"}, openai.ErrReasoningEffortUnsupported},
				{"max tokens", openai.ChatCompletionRequest{MaxTokens: 20}, openai.ErrReasoningModelMaxTokensDeprecated},
				{"temperature one", openai.ChatCompletionRequest{Temperature: 1}, openai.ErrReasoningSamplingUnsupported},
				{"temperature custom", openai.ChatCompletionRequest{Temperature: 0.7}, openai.ErrReasoningSamplingUnsupported},
				{"top p", openai.ChatCompletionRequest{TopP: 1}, openai.ErrReasoningSamplingUnsupported},
				{"logprobs", openai.ChatCompletionRequest{LogProbs: true}, openai.ErrReasoningSamplingUnsupported},
				{"top logprobs", openai.ChatCompletionRequest{TopLogProbs: 2}, openai.ErrReasoningSamplingUnsupported},
				{"tools", openai.ChatCompletionRequest{Tools: []openai.Tool{{Type: openai.ToolTypeFunction}}},
					openai.ErrReasoningToolsRequireResponses},
				{"functions", openai.ChatCompletionRequest{Functions: []openai.FunctionDefinition{{Name: "lookup"}}},
					openai.ErrReasoningToolsRequireResponses},
				{"tool choice", openai.ChatCompletionRequest{ToolChoice: "auto"}, openai.ErrReasoningToolsRequireResponses},
				{"function call", openai.ChatCompletionRequest{FunctionCall: "auto"}, openai.ErrReasoningToolsRequireResponses},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tt.request.Model = model.name
					if err := validator.Validate(tt.request); !errors.Is(err, tt.want) {
						t.Errorf("got %v, want %v", err, tt.want)
					}
				})
			}
			request := openai.ChatCompletionRequest{
				Model: model.name, ReasoningEffort: "none", Temperature: 0.7, TopP: 0.9,
				LogProbs: true, TopLogProbs: 2, Tools: []openai.Tool{{Type: openai.ToolTypeFunction}},
			}
			if err := validator.Validate(openai.ChatCompletionRequest{
				Model: model.name, ToolChoice: "none", FunctionCall: "none", ParallelToolCalls: false,
			}); err != nil {
				t.Errorf("explicitly disabling tools should not count as tool calling: %v", err)
			}
			err := validator.Validate(request)
			if model.none && err != nil {
				t.Errorf("non-reasoning request should be supported: %v", err)
			} else if !model.none && !errors.Is(err, openai.ErrReasoningEffortUnsupported) {
				t.Errorf("got %v, want unsupported reasoning effort", err)
			}
		})
	}
}

func TestModernModelsRejectLegacyCompletions(t *testing.T) {
	client := openai.NewClient("test")
	for _, model := range []string{
		openai.GPT6Dot1Sol, openai.GPT6Astra, openai.GPT6Sol, openai.GPT6Luna,
		openai.GPT5Dot6, openai.GPT5Dot6Sol, openai.GPT5Dot6Terra, openai.GPT5Dot6Luna,
	} {
		t.Run(model, func(t *testing.T) {
			request := openai.CompletionRequest{Model: model, Prompt: "Hello"}
			_, err := client.CreateCompletion(context.Background(), request)
			if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
				t.Errorf("completion: %v", err)
			}
			_, err = client.CreateCompletionStream(context.Background(), request)
			if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
				t.Errorf("stream: %v", err)
			}
		})
	}
}

func TestGPT6ValidationAppliesToBothChatMethods(t *testing.T) {
	client := openai.NewClient("test")
	request := openai.ChatCompletionRequest{Model: openai.GPT6Dot1Sol, ReasoningEffort: "none"}
	_, err := client.CreateChatCompletion(context.Background(), request)
	if !errors.Is(err, openai.ErrReasoningEffortUnsupported) {
		t.Errorf("chat: %v", err)
	}
	_, err = client.CreateChatCompletionStream(context.Background(), request)
	if !errors.Is(err, openai.ErrReasoningEffortUnsupported) {
		t.Errorf("stream: %v", err)
	}
}

func TestNewerAndCustomModelsAvoidLegacyValidation(t *testing.T) {
	for _, model := range []string{
		"my-gpt-6-deployment", "gpt-6-future", openai.GPT5Dot6,
		openai.GPT5Dot6Sol, openai.GPT5Dot6Terra, openai.GPT5Dot6Luna,
	} {
		err := openai.NewReasoningValidator().Validate(openai.ChatCompletionRequest{
			Model: model, Temperature: 0.7, ReasoningEffort: openai.ReasoningEffortNone,
		})
		if err != nil {
			t.Errorf("model %q should not inherit original GPT-5 restrictions: %v", model, err)
		}
	}
}
