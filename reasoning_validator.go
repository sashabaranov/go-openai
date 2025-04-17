package openai

import (
	"errors"
	"strings"
)

var (
	// Deprecated: use ErrReasoningModelMaxTokensDeprecated instead.
	ErrO1MaxTokensDeprecated                   = errors.New("this model is not supported MaxTokens, please use MaxCompletionTokens")                               //nolint:lll
	ErrCompletionUnsupportedModel              = errors.New("this model is not supported with this method, please use CreateChatCompletion client method instead") //nolint:lll
	ErrCompletionStreamNotSupported            = errors.New("streaming is not supported with this method, please use CreateCompletionStream")                      //nolint:lll
	ErrCompletionRequestPromptTypeNotSupported = errors.New("the type of CompletionRequest.Prompt only supports string and []string")                              //nolint:lll
)

var (
	ErrO1BetaLimitationsMessageTypes = errors.New("this model has beta-limitations, user and assistant messages only, system messages are not supported")       //nolint:lll
	ErrO1BetaLimitationsTools        = errors.New("this model has beta-limitations, tools, function calling, and response format parameters are not supported") //nolint:lll
	// Deprecated: use ErrReasoningModelLimitations* instead.
	ErrO1BetaLimitationsLogprobs = errors.New("this model has beta-limitations, logprobs not supported")                                                                               //nolint:lll
	ErrO1BetaLimitationsOther    = errors.New("this model has beta-limitations, temperature, top_p and n are fixed at 1, while presence_penalty and frequency_penalty are fixed at 0") //nolint:lll
)

var (
	//nolint:lll
	ErrReasoningModelMaxTokensDeprecated = errors.New("this model is not supported MaxTokens, please use MaxCompletionTokens")
	ErrReasoningModelLimitationsLogprobs = errors.New("this model has beta-limitations, logprobs not supported")                                                                               //nolint:lll
	ErrReasoningModelLimitationsOther    = errors.New("this model has beta-limitations, temperature, top_p and n are fixed at 1, while presence_penalty and frequency_penalty are fixed at 0") //nolint:lll
)

// ReasoningValidator handles validation for o-series model requests.
type ReasoningValidator struct{}

// NewReasoningValidator creates a new validator for o-series models.
func NewReasoningValidator() *ReasoningValidator {
	return &ReasoningValidator{}
}

// Validate performs all validation checks for o-series models.
func (v *ReasoningValidator) Validate(request ChatCompletionRequest) error {
	o1Series := strings.HasPrefix(request.Model, "o1")
	o3Series := strings.HasPrefix(request.Model, "o3")
	o4Series := strings.HasPrefix(request.Model, "o4")

	if !o1Series && !o3Series && !o4Series {
		return nil
	}

	if err := v.validateReasoningModelParams(request); err != nil {
		return err
	}

	return nil
}

// validateReasoningModelParams checks reasoning model parameters.
func (v *ReasoningValidator) validateReasoningModelParams(request ChatCompletionRequest) error {
	if request.MaxTokens > 0 {
		return ErrReasoningModelMaxTokensDeprecated
	}
	if request.LogProbs {
		return ErrReasoningModelLimitationsLogprobs
	}
	if request.Temperature > 0 && request.Temperature != 1 {
		return ErrReasoningModelLimitationsOther
	}
	if request.TopP > 0 && request.TopP != 1 {
		return ErrReasoningModelLimitationsOther
	}
	if request.N > 0 && request.N != 1 {
		return ErrReasoningModelLimitationsOther
	}
	if request.PresencePenalty > 0 {
		return ErrReasoningModelLimitationsOther
	}
	if request.FrequencyPenalty > 0 {
		return ErrReasoningModelLimitationsOther
	}

	return nil
}
