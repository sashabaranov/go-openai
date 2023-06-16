package openai

import (
	"context"
	"errors"
	"net/http"
)

// Chat message role defined by the OpenAI API.
const (
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleFunction  = "function"
)

const chatCompletionsSuffix = "/chat/completions"

var (
	ErrChatCompletionInvalidModel       = errors.New("this model is not supported with this method, please use CreateCompletion client method instead") //nolint:lll
	ErrChatCompletionStreamNotSupported = errors.New("streaming is not supported with this method, please use CreateChatCompletionStream")              //nolint:lll
)

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`

	// This property isn't in the official documentation, but it's in
	// the documentation for the official library for python:
	// - https://github.com/openai/openai-python/blob/main/chatml.md
	// - https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	Name string `json:"name,omitempty"`

	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

type FunctionCall struct {
	Name string `json:"name,omitempty"`
	// call function with arguments in JSON format
	Arguments string `json:"arguments,omitempty"`
}

// ChatCompletionRequest represents a request structure for chat completion API.
type ChatCompletionRequest struct {
	Model            string                  `json:"model"`
	Messages         []ChatCompletionMessage `json:"messages"`
	MaxTokens        int                     `json:"max_tokens,omitempty"`
	Temperature      float32                 `json:"temperature,omitempty"`
	TopP             float32                 `json:"top_p,omitempty"`
	N                int                     `json:"n,omitempty"`
	Stream           bool                    `json:"stream,omitempty"`
	Stop             []string                `json:"stop,omitempty"`
	PresencePenalty  float32                 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32                 `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]int          `json:"logit_bias,omitempty"`
	User             string                  `json:"user,omitempty"`
	Functions        []*FunctionDefine       `json:"functions,omitempty"`
	FunctionCall     string                  `json:"function_call,omitempty"`
}

type FunctionDefine struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// it's required in function call
	Parameters *FunctionParams `json:"parameters"`
}

type FunctionParams struct {
	// the Type must be JSONSchemaTypeObject
	Type       JSONSchemaType               `json:"type"`
	Properties map[string]*JSONSchemaDefine `json:"properties,omitempty"`
	Required   []string                     `json:"required,omitempty"`
}

type JSONSchemaType string

const (
	JSONSchemaTypeObject  JSONSchemaType = "object"
	JSONSchemaTypeNumber  JSONSchemaType = "number"
	JSONSchemaTypeString  JSONSchemaType = "string"
	JSONSchemaTypeArray   JSONSchemaType = "array"
	JSONSchemaTypeNull    JSONSchemaType = "null"
	JSONSchemaTypeBoolean JSONSchemaType = "boolean"
)

// JSONSchemaDefine is a struct for JSON Schema.
type JSONSchemaDefine struct {
	// Type is a type of JSON Schema.
	Type JSONSchemaType `json:"type,omitempty"`
	// Description is a description of JSON Schema.
	Description string `json:"description,omitempty"`
	// Enum is a enum of JSON Schema. It used if Type is JSONSchemaTypeString.
	Enum []string `json:"enum,omitempty"`
	// Properties is a properties of JSON Schema. It used if Type is JSONSchemaTypeObject.
	Properties map[string]*JSONSchemaDefine `json:"properties,omitempty"`
	// Required is a required of JSON Schema. It used if Type is JSONSchemaTypeObject.
	Required []string `json:"required,omitempty"`
}

type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonFunctionCall  FinishReason = "function_call"
	FinishReasonContentFilter FinishReason = "content_filter"
	FinishReasonNull          FinishReason = "null"
)

type ChatCompletionChoice struct {
	Index   int                   `json:"index"`
	Message ChatCompletionMessage `json:"message"`
	// FinishReason
	// stop: API returned complete message,
	// or a message terminated by one of the stop sequences provided via the stop parameter
	// length: Incomplete model output due to max_tokens parameter or token limit
	// function_call: The model decided to call a function
	// content_filter: Omitted content due to a flag from our content filters
	// null: API response still in progress or incomplete
	FinishReason FinishReason `json:"finish_reason"`
}

// ChatCompletionResponse represents a response structure for chat completion API.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
}

// CreateChatCompletion — API call to Create a completion for the chat message.
func (c *Client) CreateChatCompletion(
	ctx context.Context,
	request ChatCompletionRequest,
) (response ChatCompletionResponse, err error) {
	if request.Stream {
		err = ErrChatCompletionStreamNotSupported
		return
	}

	urlSuffix := chatCompletionsSuffix
	if !checkEndpointSupportsModel(urlSuffix, request.Model) {
		err = ErrChatCompletionInvalidModel
		return
	}

	req, err := c.requestBuilder.Build(ctx, http.MethodPost, c.fullURL(urlSuffix, request.Model), request)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
