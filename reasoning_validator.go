package openai

import "strings"

// ReasoningValidator handles validation for o-series model requests.
type ReasoningValidator struct {
	request ChatCompletionRequest
}

// NewReasoningValidator creates a new validator for o-series models.
func NewReasoningValidator(req ChatCompletionRequest) *ReasoningValidator {
	return &ReasoningValidator{
		request: req,
	}
}

// Validate performs all validation checks for o-series models.
func (v *ReasoningValidator) Validate() error {
	o1Series := strings.HasPrefix(v.request.Model, "o1")
	o3Series := strings.HasPrefix(v.request.Model, "o3")

	if !o1Series && !o3Series {
		return nil
	}

	if err := v.validateReasoningModelParams(); err != nil {
		return err
	}

	if o1Series {
		if err := v.validateO1Specific(); err != nil {
			return err
		}
	}

	return nil
}

// validateReasoningModelParams checks reasoning model parameters.
func (v *ReasoningValidator) validateReasoningModelParams() error {
	if v.request.MaxTokens > 0 {
		return ErrReasoningModelMaxTokensDeprecated
	}
	if v.request.LogProbs {
		return ErrReasoningModelLimitationsLogprobs
	}
	if v.request.Temperature > 0 && v.request.Temperature != 1 {
		return ErrReasoningModelLimitationsOther
	}
	if v.request.TopP > 0 && v.request.TopP != 1 {
		return ErrReasoningModelLimitationsOther
	}
	if v.request.N > 0 && v.request.N != 1 {
		return ErrReasoningModelLimitationsOther
	}
	if v.request.PresencePenalty > 0 {
		return ErrReasoningModelLimitationsOther
	}
	if v.request.FrequencyPenalty > 0 {
		return ErrReasoningModelLimitationsOther
	}

	return nil
}

// validateO1Specific checks O1-specific limitations.
func (v *ReasoningValidator) validateO1Specific() error {
	for _, m := range v.request.Messages {
		if _, found := availableMessageRoleForO1Models[m.Role]; !found {
			return ErrO1BetaLimitationsMessageTypes
		}
	}

	for _, t := range v.request.Tools {
		if _, found := unsupportedToolsForO1Models[t.Type]; found {
			return ErrO1BetaLimitationsTools
		}
	}
	return nil
}
