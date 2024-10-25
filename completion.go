package openai

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrO1MaxTokensDeprecated                   = errors.New("this model is not supported MaxTokens, please use MaxCompletionTokens")                               //nolint:lll
	ErrCompletionUnsupportedModel              = errors.New("this model is not supported with this method, please use CreateChatCompletion client method instead") //nolint:lll
	ErrCompletionStreamNotSupported            = errors.New("streaming is not supported with this method, please use CreateCompletionStream")                      //nolint:lll
	ErrCompletionRequestPromptTypeNotSupported = errors.New("the type of CompletionRequest.Prompt only supports string and []string")                              //nolint:lll
)

var (
	ErrO1BetaLimitationsMessageTypes = errors.New("this model has beta-limitations, user and assistant messages only, system messages are not supported")                                  //nolint:lll
	ErrO1BetaLimitationsStreaming    = errors.New("this model has beta-limitations, streaming not supported")                                                                              //nolint:lll
	ErrO1BetaLimitationsTools        = errors.New("this model has beta-limitations, tools, function calling, and response format parameters are not supported")                            //nolint:lll
	ErrO1BetaLimitationsLogprobs     = errors.New("this model has beta-limitations, logprobs not supported")                                                                               //nolint:lll
	ErrO1BetaLimitationsOther        = errors.New("this model has beta-limitations, temperature, top_p and n are fixed at 1, while presence_penalty and frequency_penalty are fixed at 0") //nolint:lll
)

// GPT3 Defines the models provided by OpenAI to use when generating
// completions from OpenAI.
// GPT3 Models are designed for text-based tasks. For code-specific
// tasks, please refer to the Codex series of models.
const (
	O1Mini                = "o1-mini"
	O1Mini20240912        = "o1-mini-2024-09-12"
	O1Preview             = "o1-preview"
	O1Preview20240912     = "o1-preview-2024-09-12"
	GPT432K0613           = "gpt-4-32k-0613"
	GPT432K0314           = "gpt-4-32k-0314"
	GPT432K               = "gpt-4-32k"
	GPT40613              = "gpt-4-0613"
	GPT40314              = "gpt-4-0314"
	GPT4o                 = "gpt-4o"
	GPT4o20240513         = "gpt-4o-2024-05-13"
	GPT4o20240806         = "gpt-4o-2024-08-06"
	GPT4oLatest           = "chatgpt-4o-latest"
	GPT4oMini             = "gpt-4o-mini"
	GPT4oMini20240718     = "gpt-4o-mini-2024-07-18"
	GPT4Turbo             = "gpt-4-turbo"
	GPT4Turbo20240409     = "gpt-4-turbo-2024-04-09"
	GPT4Turbo0125         = "gpt-4-0125-preview"
	GPT4Turbo1106         = "gpt-4-1106-preview"
	GPT4TurboPreview      = "gpt-4-turbo-preview"
	GPT4VisionPreview     = "gpt-4-vision-preview"
	GPT4                  = "gpt-4"
	GPT3Dot5Turbo0125     = "gpt-3.5-turbo-0125"
	GPT3Dot5Turbo1106     = "gpt-3.5-turbo-1106"
	GPT3Dot5Turbo0613     = "gpt-3.5-turbo-0613"
	GPT3Dot5Turbo0301     = "gpt-3.5-turbo-0301"
	GPT3Dot5Turbo16K      = "gpt-3.5-turbo-16k"
	GPT3Dot5Turbo16K0613  = "gpt-3.5-turbo-16k-0613"
	GPT3Dot5Turbo         = "gpt-3.5-turbo"
	GPT3Dot5TurboInstruct = "gpt-3.5-turbo-instruct"
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	GPT3TextDavinci003 = "text-davinci-003"
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	GPT3TextDavinci002 = "text-davinci-002"
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	GPT3TextCurie001 = "text-curie-001"
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	GPT3TextBabbage001 = "text-babbage-001"
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	GPT3TextAda001 = "text-ada-001"
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	GPT3TextDavinci001 = "text-davinci-001"
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	GPT3DavinciInstructBeta = "davinci-instruct-beta"
	// Deprecated: Model is shutdown. Use davinci-002 instead.
	GPT3Davinci    = "davinci"
	GPT3Davinci002 = "davinci-002"
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	GPT3CurieInstructBeta = "curie-instruct-beta"
	GPT3Curie             = "curie"
	GPT3Curie002          = "curie-002"
	// Deprecated: Model is shutdown. Use babbage-002 instead.
	GPT3Ada    = "ada"
	GPT3Ada002 = "ada-002"
	// Deprecated: Model is shutdown. Use babbage-002 instead.
	GPT3Babbage    = "babbage"
	GPT3Babbage002 = "babbage-002"
)

// Codex Defines the models provided by OpenAI.
// These models are designed for code-specific tasks, and use
// a different tokenizer which optimizes for whitespace.
const (
	CodexCodeDavinci002 = "code-davinci-002"
	CodexCodeCushman001 = "code-cushman-001"
	CodexCodeDavinci001 = "code-davinci-001"
)

// O1SeriesModels List of new Series of OpenAI models.
// Some old api attributes not supported.
var O1SeriesModels = map[string]struct{}{
	O1Mini:            {},
	O1Mini20240912:    {},
	O1Preview:         {},
	O1Preview20240912: {},
}

var disabledModelsForEndpoints = map[string]map[string]bool{
	"/completions": {
		O1Mini:               true,
		O1Mini20240912:       true,
		O1Preview:            true,
		O1Preview20240912:    true,
		GPT3Dot5Turbo:        true,
		GPT3Dot5Turbo0301:    true,
		GPT3Dot5Turbo0613:    true,
		GPT3Dot5Turbo1106:    true,
		GPT3Dot5Turbo0125:    true,
		GPT3Dot5Turbo16K:     true,
		GPT3Dot5Turbo16K0613: true,
		GPT4:                 true,
		GPT4o:                true,
		GPT4o20240513:        true,
		GPT4o20240806:        true,
		GPT4oLatest:          true,
		GPT4oMini:            true,
		GPT4oMini20240718:    true,
		GPT4TurboPreview:     true,
		GPT4VisionPreview:    true,
		GPT4Turbo1106:        true,
		GPT4Turbo0125:        true,
		GPT4Turbo:            true,
		GPT4Turbo20240409:    true,
		GPT40314:             true,
		GPT40613:             true,
		GPT432K:              true,
		GPT432K0314:          true,
		GPT432K0613:          true,
	},
	chatCompletionsSuffix: {
		CodexCodeDavinci002:     true,
		CodexCodeCushman001:     true,
		CodexCodeDavinci001:     true,
		GPT3TextDavinci003:      true,
		GPT3TextDavinci002:      true,
		GPT3TextCurie001:        true,
		GPT3TextBabbage001:      true,
		GPT3TextAda001:          true,
		GPT3TextDavinci001:      true,
		GPT3DavinciInstructBeta: true,
		GPT3Davinci:             true,
		GPT3CurieInstructBeta:   true,
		GPT3Curie:               true,
		GPT3Ada:                 true,
		GPT3Babbage:             true,
	},
}

func checkEndpointSupportsModel(endpoint, model string) bool {
	return !disabledModelsForEndpoints[endpoint][model]
}

func checkPromptType(prompt any) bool {
	_, isString := prompt.(string)
	_, isStringSlice := prompt.([]string)
	if isString || isStringSlice {
		return true
	}

	// check if it is prompt is []string hidden under []any
	slice, isSlice := prompt.([]any)
	if !isSlice {
		return false
	}

	for _, item := range slice {
		_, itemIsString := item.(string)
		if !itemIsString {
			return false
		}
	}
	return true // all items in the slice are string, so it is []string
}

var unsupportedToolsForO1Models = map[ToolType]struct{}{
	ToolTypeFunction: {},
}

var availableMessageRoleForO1Models = map[string]struct{}{
	ChatMessageRoleUser:      {},
	ChatMessageRoleAssistant: {},
}

// validateRequestForO1Models checks for deprecated fields of OpenAI models.
func validateRequestForO1Models(request ChatCompletionRequest) error {
	if _, found := O1SeriesModels[request.Model]; !found {
		return nil
	}

	if request.MaxTokens > 0 {
		return ErrO1MaxTokensDeprecated
	}

	// Beta Limitations
	// refs:https://platform.openai.com/docs/guides/reasoning/beta-limitations
	// Streaming: not supported
	if request.Stream {
		return ErrO1BetaLimitationsStreaming
	}
	// Logprobs: not supported.
	if request.LogProbs {
		return ErrO1BetaLimitationsLogprobs
	}

	// Message types: user and assistant messages only, system messages are not supported.
	for _, m := range request.Messages {
		if _, found := availableMessageRoleForO1Models[m.Role]; !found {
			return ErrO1BetaLimitationsMessageTypes
		}
	}

	// Tools: tools, function calling, and response format parameters are not supported
	for _, t := range request.Tools {
		if _, found := unsupportedToolsForO1Models[t.Type]; found {
			return ErrO1BetaLimitationsTools
		}
	}

	// Other: temperature, top_p and n are fixed at 1, while presence_penalty and frequency_penalty are fixed at 0.
	if request.Temperature > 0 && request.Temperature != 1 {
		return ErrO1BetaLimitationsOther
	}
	if request.TopP > 0 && request.TopP != 1 {
		return ErrO1BetaLimitationsOther
	}
	if request.N > 0 && request.N != 1 {
		return ErrO1BetaLimitationsOther
	}
	if request.PresencePenalty > 0 {
		return ErrO1BetaLimitationsOther
	}
	if request.FrequencyPenalty > 0 {
		return ErrO1BetaLimitationsOther
	}

	return nil
}

// CompletionRequest represents a request structure for completion API.
type CompletionRequest struct {
	Model            string  `json:"model"`
	Prompt           any     `json:"prompt,omitempty"`
	BestOf           int     `json:"best_of,omitempty"`
	Echo             bool    `json:"echo,omitempty"`
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`
	// LogitBias is must be a token id string (specified by their token ID in the tokenizer), not a word string.
	// incorrect: `"logit_bias":{"You": 6}`, correct: `"logit_bias":{"1639": 6}`
	// refs: https://platform.openai.com/docs/api-reference/completions/create#completions/create-logit_bias
	LogitBias map[string]int `json:"logit_bias,omitempty"`
	// Store can be set to true to store the output of this completion request for use in distillations and evals.
	// https://platform.openai.com/docs/api-reference/chat/create#chat-create-store
	Store bool `json:"store,omitempty"`
	// Metadata to store with the completion.
	Metadata        map[string]string `json:"metadata,omitempty"`
	LogProbs        int               `json:"logprobs,omitempty"`
	MaxTokens       int               `json:"max_tokens,omitempty"`
	N               int               `json:"n,omitempty"`
	PresencePenalty float32           `json:"presence_penalty,omitempty"`
	Seed            *int              `json:"seed,omitempty"`
	Stop            []string          `json:"stop,omitempty"`
	Stream          bool              `json:"stream,omitempty"`
	Suffix          string            `json:"suffix,omitempty"`
	Temperature     float32           `json:"temperature,omitempty"`
	TopP            float32           `json:"top_p,omitempty"`
	User            string            `json:"user,omitempty"`
}

// CompletionChoice represents one of possible completions.
type CompletionChoice struct {
	Text         string        `json:"text"`
	Index        int           `json:"index"`
	FinishReason string        `json:"finish_reason"`
	LogProbs     LogprobResult `json:"logprobs"`
}

// LogprobResult represents logprob result of Choice.
type LogprobResult struct {
	Tokens        []string             `json:"tokens"`
	TokenLogprobs []float32            `json:"token_logprobs"`
	TopLogprobs   []map[string]float32 `json:"top_logprobs"`
	TextOffset    []int                `json:"text_offset"`
}

// CompletionResponse represents a response structure for completion API.
type CompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   Usage              `json:"usage"`

	httpHeader
}

// CreateCompletion — API call to create a completion. This is the main endpoint of the API. Returns new text as well
// as, if requested, the probabilities over each alternative token at each position.
//
// If using a fine-tuned model, simply provide the model's ID in the CompletionRequest object,
// and the server will use the model's parameters to generate the completion.
func (c *Client) CreateCompletion(
	ctx context.Context,
	request CompletionRequest,
) (response CompletionResponse, err error) {
	if request.Stream {
		err = ErrCompletionStreamNotSupported
		return
	}

	urlSuffix := "/completions"
	if !checkEndpointSupportsModel(urlSuffix, request.Model) {
		err = ErrCompletionUnsupportedModel
		return
	}

	if !checkPromptType(request.Prompt) {
		err = ErrCompletionRequestPromptTypeNotSupported
		return
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix, withModel(request.Model)),
		withBody(request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
