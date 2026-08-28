package openai

import (
	"context"
	"net/http"
)

const completionsSuffix = "/completions"

// Text generation and reasoning models provided by OpenAI.
const (
	O1Mini                  = "o1-mini"
	O1Mini20240912          = "o1-mini-2024-09-12"
	O1Preview               = "o1-preview"
	O1Preview20240912       = "o1-preview-2024-09-12"
	O1                      = "o1"
	O120241217              = "o1-2024-12-17"
	O3                      = "o3"
	O320250416              = "o3-2025-04-16"
	O3Mini                  = "o3-mini"
	O3Mini20250131          = "o3-mini-2025-01-31"
	O3Pro                   = "o3-pro"
	O3DeepResearch          = "o3-deep-research"
	O4Mini                  = "o4-mini"
	O4Mini20250416          = "o4-mini-2025-04-16"
	O4MiniDeepResearch      = "o4-mini-deep-research"
	GPT432K0613             = "gpt-4-32k-0613"
	GPT432K0314             = "gpt-4-32k-0314"
	GPT432K                 = "gpt-4-32k"
	GPT40613                = "gpt-4-0613"
	GPT40314                = "gpt-4-0314"
	GPT4o                   = "gpt-4o"
	GPT4o20240513           = "gpt-4o-2024-05-13"
	GPT4o20240806           = "gpt-4o-2024-08-06"
	GPT4o20241120           = "gpt-4o-2024-11-20"
	GPT4oLatest             = "chatgpt-4o-latest"
	GPT4oMini               = "gpt-4o-mini"
	GPT4oMini20240718       = "gpt-4o-mini-2024-07-18"
	GPT4Turbo               = "gpt-4-turbo"
	GPT4Turbo20240409       = "gpt-4-turbo-2024-04-09"
	GPT4Turbo0125           = "gpt-4-0125-preview"
	GPT4Turbo1106           = "gpt-4-1106-preview"
	GPT4TurboPreview        = "gpt-4-turbo-preview"
	GPT4VisionPreview       = "gpt-4-vision-preview"
	GPT4                    = "gpt-4"
	GPT4Dot1                = "gpt-4.1"
	GPT4Dot120250414        = "gpt-4.1-2025-04-14"
	GPT4Dot1Mini            = "gpt-4.1-mini"
	GPT4Dot1Mini20250414    = "gpt-4.1-mini-2025-04-14"
	GPT4Dot1Nano            = "gpt-4.1-nano"
	GPT4Dot1Nano20250414    = "gpt-4.1-nano-2025-04-14"
	GPT4Dot5Preview         = "gpt-4.5-preview"
	GPT4Dot5Preview20250227 = "gpt-4.5-preview-2025-02-27"
	GPT5                    = "gpt-5"
	GPT5Mini                = "gpt-5-mini"
	GPT5Nano                = "gpt-5-nano"
	GPT5Pro                 = "gpt-5-pro"
	GPT5ChatLatest          = "gpt-5-chat-latest"
	GPT5Codex               = "gpt-5-codex"
	GPT5Dot1                = "gpt-5.1"
	GPT5Dot1ChatLatest      = "gpt-5.1-chat-latest"
	GPT5Dot1Codex           = "gpt-5.1-codex"
	GPT5Dot1CodexMini       = "gpt-5.1-codex-mini"
	GPT5Dot1CodexMax        = "gpt-5.1-codex-max"
	GPT5Dot2                = "gpt-5.2"
	GPT5Dot2ChatLatest      = "gpt-5.2-chat-latest"
	GPT5Dot2Pro             = "gpt-5.2-pro"
	GPT5Dot2Codex           = "gpt-5.2-codex"
	GPT5Dot3ChatLatest      = "gpt-5.3-chat-latest"
	GPT5Dot3Codex           = "gpt-5.3-codex"
	GPT5Dot4                = "gpt-5.4"
	GPT5Dot4Mini            = "gpt-5.4-mini"
	GPT5Dot4Nano            = "gpt-5.4-nano"
	GPT5Dot4Pro             = "gpt-5.4-pro"
	GPT5Dot5                = "gpt-5.5"
	GPT5Dot5Pro             = "gpt-5.5-pro"
	GPT5Dot6                = "gpt-5.6"
	GPT5Dot6Sol             = "gpt-5.6-sol"
	GPT5Dot6Terra           = "gpt-5.6-terra"
	GPT5Dot6Luna            = "gpt-5.6-luna"
	GPT3Dot5Turbo0125       = "gpt-3.5-turbo-0125"
	GPT3Dot5Turbo1106       = "gpt-3.5-turbo-1106"
	GPT3Dot5Turbo0613       = "gpt-3.5-turbo-0613"
	GPT3Dot5Turbo0301       = "gpt-3.5-turbo-0301"
	GPT3Dot5Turbo16K        = "gpt-3.5-turbo-16k"
	GPT3Dot5Turbo16K0613    = "gpt-3.5-turbo-16k-0613"
	GPT3Dot5Turbo           = "gpt-3.5-turbo"
	GPT3Dot5TurboInstruct   = "gpt-3.5-turbo-instruct"
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

var disabledModelsForEndpoints = map[string]map[string]bool{
	completionsSuffix: {
		O1Mini:                  true,
		O1Mini20240912:          true,
		O1Preview:               true,
		O1Preview20240912:       true,
		O3Mini:                  true,
		O3Mini20250131:          true,
		O3Pro:                   true,
		O3DeepResearch:          true,
		O4Mini:                  true,
		O4Mini20250416:          true,
		O4MiniDeepResearch:      true,
		O3:                      true,
		O320250416:              true,
		GPT3Dot5Turbo:           true,
		GPT3Dot5Turbo0301:       true,
		GPT3Dot5Turbo0613:       true,
		GPT3Dot5Turbo1106:       true,
		GPT3Dot5Turbo0125:       true,
		GPT3Dot5Turbo16K:        true,
		GPT3Dot5Turbo16K0613:    true,
		GPT4:                    true,
		GPT4Dot5Preview:         true,
		GPT4Dot5Preview20250227: true,
		GPT4o:                   true,
		GPT4o20240513:           true,
		GPT4o20240806:           true,
		GPT4o20241120:           true,
		GPT4oLatest:             true,
		GPT4oMini:               true,
		GPT4oMini20240718:       true,
		GPT4TurboPreview:        true,
		GPT4VisionPreview:       true,
		GPT4Turbo1106:           true,
		GPT4Turbo0125:           true,
		GPT4Turbo:               true,
		GPT4Turbo20240409:       true,
		GPT40314:                true,
		GPT40613:                true,
		GPT432K:                 true,
		GPT432K0314:             true,
		GPT432K0613:             true,
		O1:                      true,
		GPT4Dot1:                true,
		GPT4Dot120250414:        true,
		GPT4Dot1Mini:            true,
		GPT4Dot1Mini20250414:    true,
		GPT4Dot1Nano:            true,
		GPT4Dot1Nano20250414:    true,
		GPT5:                    true,
		GPT5Mini:                true,
		GPT5Nano:                true,
		GPT5Pro:                 true,
		GPT5ChatLatest:          true,
		GPT5Codex:               true,
		GPT5Dot1:                true,
		GPT5Dot1ChatLatest:      true,
		GPT5Dot1Codex:           true,
		GPT5Dot1CodexMini:       true,
		GPT5Dot1CodexMax:        true,
		GPT5Dot2:                true,
		GPT5Dot2ChatLatest:      true,
		GPT5Dot2Pro:             true,
		GPT5Dot2Codex:           true,
		GPT5Dot3ChatLatest:      true,
		GPT5Dot3Codex:           true,
		GPT5Dot4:                true,
		GPT5Dot4Mini:            true,
		GPT5Dot4Nano:            true,
		GPT5Dot4Pro:             true,
		GPT5Dot5:                true,
		GPT5Dot5Pro:             true,
		GPT5Dot6:                true,
		GPT5Dot6Sol:             true,
		GPT5Dot6Terra:           true,
		GPT5Dot6Luna:            true,
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
	Stream          bool              `json:"stream"`
	Suffix          string            `json:"suffix,omitempty"`
	Temperature     float32           `json:"temperature,omitempty"`
	TopP            float32           `json:"top_p,omitempty"`
	User            string            `json:"user,omitempty"`
	// Options for streaming response. Only set this when you set stream: true.
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
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
	Usage   *Usage             `json:"usage,omitempty"`

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

	urlSuffix := completionsSuffix
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
