package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const responsesSuffix = "/responses"

var ErrResponseStreamNotSupported = errors.New(
	"streaming is not supported with this method, please use CreateResponseStream",
)

// ResponseInclude identifies optional data to include in a response.
type ResponseInclude string

const (
	ResponseIncludeFileSearchCallResults      ResponseInclude = "file_search_call.results"
	ResponseIncludeWebSearchCallResults       ResponseInclude = "web_search_call.results"
	ResponseIncludeWebSearchCallActionSources ResponseInclude = "web_search_call.action.sources"
	ResponseIncludeInputImageURL              ResponseInclude = "message.input_image.image_url"
	ResponseIncludeComputerCallOutputImageURL ResponseInclude = "computer_call_output.output.image_url"
	ResponseIncludeCodeInterpreterCallOutputs ResponseInclude = "code_interpreter_call.outputs"
	ResponseIncludeReasoningEncryptedContent  ResponseInclude = "reasoning.encrypted_content"
	ResponseIncludeMessageOutputTextLogprobs  ResponseInclude = "message.output_text.logprobs"
)

// ResponseStatus is the lifecycle status of a response.
type ResponseStatus string

const (
	ResponseStatusQueued     ResponseStatus = "queued"
	ResponseStatusInProgress ResponseStatus = "in_progress"
	ResponseStatusCompleted  ResponseStatus = "completed"
	ResponseStatusFailed     ResponseStatus = "failed"
	ResponseStatusIncomplete ResponseStatus = "incomplete"
	ResponseStatusCancelling ResponseStatus = "cancelling"
	ResponseStatusCancelled  ResponseStatus = "cancelled"
)

// ResponseTruncation controls how input that exceeds the context window is handled.
type ResponseTruncation string

const (
	ResponseTruncationAuto     ResponseTruncation = "auto"
	ResponseTruncationDisabled ResponseTruncation = "disabled"
)

// ResponseTool is an alias for Tool. Responses API tool-specific properties can
// be supplied through Tool.Parameters.
type ResponseTool = Tool

// NewResponseFunctionTool converts a function definition to the inline function
// tool representation expected by the Responses API.
func NewResponseFunctionTool(function FunctionDefinition) ResponseTool {
	parameters := map[string]any{
		"name":       function.Name,
		"parameters": function.Parameters,
	}
	if function.Description != "" {
		parameters["description"] = function.Description
	}
	if function.Strict {
		parameters["strict"] = true
	}
	return ResponseTool{Type: ToolTypeFunction, Parameters: parameters}
}

// ResponseReasoning represents reasoning configuration for the Responses API.
type ResponseReasoning struct {
	Effort          string `json:"effort,omitempty"`
	GenerateSummary string `json:"generate_summary,omitempty"`
	Summary         string `json:"summary,omitempty"`
	Context         string `json:"context,omitempty"`
	Mode            string `json:"mode,omitempty"`
}

// ResponseStreamOptions controls Responses API streaming behavior.
type ResponseStreamOptions struct {
	IncludeObfuscation *bool `json:"include_obfuscation,omitempty"`
}

// ResponseTextConfig controls plain-text or structured response output.
type ResponseTextConfig struct {
	Format    *ResponseTextFormat `json:"format,omitempty"`
	Verbosity string              `json:"verbosity,omitempty"`
}

// ResponseTextFormat describes the requested output format.
type ResponseTextFormat struct {
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Schema      any    `json:"schema,omitempty"`
	Strict      bool   `json:"strict,omitempty"`
}

// ResponsePrompt references a reusable prompt template.
type ResponsePrompt struct {
	ID        string         `json:"id"`
	Variables map[string]any `json:"variables,omitempty"`
	Version   string         `json:"version,omitempty"`
}

// ResponsePromptCacheOptions controls prompt cache creation.
type ResponsePromptCacheOptions struct {
	Mode string `json:"mode,omitempty"`
	TTL  string `json:"ttl,omitempty"`
}

// ResponsePromptCacheBreakpoint marks the end of a reusable prompt prefix.
type ResponsePromptCacheBreakpoint struct {
	Mode string `json:"mode"`
}

// CreateResponseRequest represents a request to the Responses API. Input may be
// a string or a slice of response input items.
type CreateResponseRequest struct {
	Background           bool                        `json:"background,omitempty"`
	ContextManagement    []any                       `json:"context_management,omitempty"`
	Conversation         any                         `json:"conversation,omitempty"`
	Include              []ResponseInclude           `json:"include,omitempty"`
	Input                any                         `json:"input"`
	Instructions         string                      `json:"instructions,omitempty"`
	MaxOutputTokens      int                         `json:"max_output_tokens,omitempty"`
	MaxToolCalls         int                         `json:"max_tool_calls,omitempty"`
	Metadata             map[string]any              `json:"metadata,omitempty"`
	Model                string                      `json:"model,omitempty"`
	Moderation           any                         `json:"moderation,omitempty"`
	ParallelToolCalls    *bool                       `json:"parallel_tool_calls,omitempty"`
	PreviousResponseID   string                      `json:"previous_response_id,omitempty"`
	Prompt               *ResponsePrompt             `json:"prompt,omitempty"`
	PromptCacheKey       string                      `json:"prompt_cache_key,omitempty"`
	PromptCacheOptions   *ResponsePromptCacheOptions `json:"prompt_cache_options,omitempty"`
	PromptCacheRetention string                      `json:"prompt_cache_retention,omitempty"`
	Reasoning            *ResponseReasoning          `json:"reasoning,omitempty"`
	SafetyIdentifier     string                      `json:"safety_identifier,omitempty"`
	ServiceTier          string                      `json:"service_tier,omitempty"`
	Store                *bool                       `json:"store,omitempty"`
	Stream               bool                        `json:"stream,omitempty"`
	StreamOptions        *ResponseStreamOptions      `json:"stream_options,omitempty"`
	Temperature          *float32                    `json:"temperature,omitempty"`
	Text                 *ResponseTextConfig         `json:"text,omitempty"`
	ToolChoice           any                         `json:"tool_choice,omitempty"`
	Tools                []ResponseTool              `json:"tools,omitempty"`
	TopLogprobs          int                         `json:"top_logprobs,omitempty"`
	TopP                 *float32                    `json:"top_p,omitempty"`
	Truncation           ResponseTruncation          `json:"truncation,omitempty"`
	User                 string                      `json:"user,omitempty"`
	ExtraBody            map[string]any              `json:"-"`
}

// MarshalJSON merges ExtraBody into the request payload. ExtraBody values take
// precedence over fields represented directly by CreateResponseRequest.
func (r CreateResponseRequest) MarshalJSON() ([]byte, error) {
	type requestAlias CreateResponseRequest
	base, err := json.Marshal(requestAlias(r))
	if err != nil || len(r.ExtraBody) == 0 {
		return base, err
	}

	var body map[string]any
	if err = json.Unmarshal(base, &body); err != nil {
		return nil, err
	}
	for key, value := range r.ExtraBody {
		body[key] = value
	}
	return json.Marshal(body)
}

// ResponseInputMessage is a message supplied as structured input.
type ResponseInputMessage struct {
	Type    string `json:"type,omitempty"`
	Role    string `json:"role"`
	Content any    `json:"content"`
	Status  string `json:"status,omitempty"`
	Phase   string `json:"phase,omitempty"`
}

// ResponseInputText is a text content part in a structured input message.
type ResponseInputText struct {
	Type                  string                         `json:"type"`
	Text                  string                         `json:"text"`
	PromptCacheBreakpoint *ResponsePromptCacheBreakpoint `json:"prompt_cache_breakpoint,omitempty"`
}

// ResponseInputImage is an image content part in a structured input message.
type ResponseInputImage struct {
	Type                  string                         `json:"type"`
	Detail                string                         `json:"detail,omitempty"`
	FileID                string                         `json:"file_id,omitempty"`
	ImageURL              string                         `json:"image_url,omitempty"`
	PromptCacheBreakpoint *ResponsePromptCacheBreakpoint `json:"prompt_cache_breakpoint,omitempty"`
}

// ResponseInputFile is a file content part in a structured input message.
type ResponseInputFile struct {
	Type                  string                         `json:"type"`
	FileData              string                         `json:"file_data,omitempty"`
	FileID                string                         `json:"file_id,omitempty"`
	FileURL               string                         `json:"file_url,omitempty"`
	Filename              string                         `json:"filename,omitempty"`
	Detail                string                         `json:"detail,omitempty"`
	PromptCacheBreakpoint *ResponsePromptCacheBreakpoint `json:"prompt_cache_breakpoint,omitempty"`
}

// ResponseFunctionCallOutput supplies the result of a prior function call.
type ResponseFunctionCallOutput struct {
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output any    `json:"output"`
	Status string `json:"status,omitempty"`
}

// ResponseError is an error embedded in an otherwise successful Responses API request.
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ResponseIncompleteDetails explains why a response did not complete.
type ResponseIncompleteDetails struct {
	Reason string `json:"reason,omitempty"`
}

// ResponseConversation identifies the conversation associated with a response.
type ResponseConversation struct {
	ID string `json:"id"`
}

// ResponseUsage reports token use for a response.
type ResponseUsage struct {
	InputTokens         int                          `json:"input_tokens"`
	InputTokensDetails  *ResponseInputTokensDetails  `json:"input_tokens_details,omitempty"`
	OutputTokens        int                          `json:"output_tokens"`
	OutputTokensDetails *ResponseOutputTokensDetails `json:"output_tokens_details,omitempty"`
	TotalTokens         int                          `json:"total_tokens"`
}

// ResponseInputTokensDetails is the input-token usage breakdown.
type ResponseInputTokensDetails struct {
	CachedTokens     int `json:"cached_tokens"`
	CacheWriteTokens int `json:"cache_write_tokens"`
}

// ResponseOutputTokensDetails is the output-token usage breakdown.
type ResponseOutputTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// ResponseAnnotation describes a citation or file annotation in output text.
type ResponseAnnotation struct {
	Type       string `json:"type"`
	FileID     string `json:"file_id,omitempty"`
	Filename   string `json:"filename,omitempty"`
	Index      int    `json:"index,omitempty"`
	StartIndex int    `json:"start_index,omitempty"`
	EndIndex   int    `json:"end_index,omitempty"`
	URL        string `json:"url,omitempty"`
	Title      string `json:"title,omitempty"`
}

// ResponseLogprob contains token log-probability information.
type ResponseLogprob struct {
	Token       string            `json:"token"`
	Bytes       []int64           `json:"bytes,omitempty"`
	Logprob     float64           `json:"logprob"`
	TopLogprobs []ResponseLogprob `json:"top_logprobs,omitempty"`
}

// ResponseOutputContent is a text or refusal content part in an output message.
type ResponseOutputContent struct {
	Type        string               `json:"type"`
	Text        string               `json:"text,omitempty"`
	Refusal     string               `json:"refusal,omitempty"`
	Annotations []ResponseAnnotation `json:"annotations,omitempty"`
	Logprobs    []ResponseLogprob    `json:"logprobs,omitempty"`
}

// ResponseSummaryPart is a reasoning summary content part.
type ResponseSummaryPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ResponseOutputItem contains the common fields shared by response output item variants.
// The top-level Output field remains []any so new variants can be consumed without a library release.
type ResponseOutputItem struct {
	ID        string                  `json:"id,omitempty"`
	Type      string                  `json:"type"`
	Status    string                  `json:"status,omitempty"`
	Role      string                  `json:"role,omitempty"`
	Content   []ResponseOutputContent `json:"content,omitempty"`
	CallID    string                  `json:"call_id,omitempty"`
	Name      string                  `json:"name,omitempty"`
	Arguments string                  `json:"arguments,omitempty"`
	Summary   []ResponseSummaryPart   `json:"summary,omitempty"`
	Action    any                     `json:"action,omitempty"`
	Results   any                     `json:"results,omitempty"`
	Output    any                     `json:"output,omitempty"`
}

// CreateResponseResponse represents a response returned by the Responses API.
type CreateResponseResponse struct {
	ID                   string                      `json:"id"`
	Object               string                      `json:"object"`
	Created              int64                       `json:"created_at"`
	CompletedAt          *int64                      `json:"completed_at,omitempty"`
	Status               ResponseStatus              `json:"status,omitempty"`
	Error                *ResponseError              `json:"error,omitempty"`
	IncompleteDetails    *ResponseIncompleteDetails  `json:"incomplete_details,omitempty"`
	Instructions         any                         `json:"instructions,omitempty"`
	MaxOutputTokens      *int                        `json:"max_output_tokens,omitempty"`
	MaxToolCalls         *int                        `json:"max_tool_calls,omitempty"`
	Metadata             map[string]any              `json:"metadata,omitempty"`
	Model                string                      `json:"model"`
	Moderation           any                         `json:"moderation,omitempty"`
	Output               []any                       `json:"output"`
	OutputText           string                      `json:"output_text,omitempty"`
	ParallelToolCalls    bool                        `json:"parallel_tool_calls,omitempty"`
	PreviousResponseID   string                      `json:"previous_response_id,omitempty"`
	Reasoning            *ResponseReasoning          `json:"reasoning,omitempty"`
	ServiceTier          string                      `json:"service_tier,omitempty"`
	Store                bool                        `json:"store,omitempty"`
	Temperature          *float32                    `json:"temperature,omitempty"`
	Text                 *ResponseTextConfig         `json:"text,omitempty"`
	ToolChoice           any                         `json:"tool_choice,omitempty"`
	Tools                []any                       `json:"tools,omitempty"`
	TopLogprobs          int                         `json:"top_logprobs,omitempty"`
	TopP                 *float32                    `json:"top_p,omitempty"`
	Truncation           ResponseTruncation          `json:"truncation,omitempty"`
	Usage                *ResponseUsage              `json:"usage,omitempty"`
	Background           *bool                       `json:"background,omitempty"`
	Conversation         *ResponseConversation       `json:"conversation,omitempty"`
	Prompt               *ResponsePrompt             `json:"prompt,omitempty"`
	PromptCacheKey       string                      `json:"prompt_cache_key,omitempty"`
	PromptCacheOptions   *ResponsePromptCacheOptions `json:"prompt_cache_options,omitempty"`
	PromptCacheRetention string                      `json:"prompt_cache_retention,omitempty"`
	SafetyIdentifier     string                      `json:"safety_identifier,omitempty"`
	User                 string                      `json:"user,omitempty"`

	httpHeader
}

// GetOutputText returns the aggregated text output. It uses the API's output_text
// convenience field when present and otherwise extracts output_text content parts.
func (r CreateResponseResponse) GetOutputText() string {
	if r.OutputText != "" {
		return r.OutputText
	}

	var output strings.Builder
	for _, rawItem := range r.Output {
		data, err := json.Marshal(rawItem)
		if err != nil {
			continue
		}
		var item ResponseOutputItem
		if err = json.Unmarshal(data, &item); err != nil {
			continue
		}
		for _, content := range item.Content {
			if content.Type == "output_text" {
				output.WriteString(content.Text)
			}
		}
	}
	return output.String()
}

// RetrieveResponseOptions controls optional data returned by RetrieveResponse.
type RetrieveResponseOptions struct {
	Include            []ResponseInclude
	IncludeObfuscation *bool
	StartingAfter      *int
}

// ResponseInputItemsListOptions controls pagination for ListResponseInputItems.
type ResponseInputItemsListOptions struct {
	After   string
	Include []ResponseInclude
	Limit   int
	Order   string
}

// ResponseInputItemsList contains the input items for a response.
type ResponseInputItemsList struct {
	Object  string `json:"object"`
	Data    []any  `json:"data"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
	HasMore bool   `json:"has_more"`

	httpHeader
}

// ResponseDeleteResponse is returned after deleting a stored response.
type ResponseDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

// DeleteResponseResponse is kept as a descriptive alias for ResponseDeleteResponse.
type DeleteResponseResponse = ResponseDeleteResponse

// ResponseInputTokensResponse reports the token count for response input.
type ResponseInputTokensResponse struct {
	Object      string `json:"object"`
	InputTokens int    `json:"input_tokens"`

	httpHeader
}

// ResponseInputTokensRequest contains the response input whose tokens should be counted.
type ResponseInputTokensRequest = CreateResponseRequest

// ResponseCompaction is a compacted response context.
type ResponseCompaction struct {
	ID        string         `json:"id"`
	Object    string         `json:"object"`
	CreatedAt int64          `json:"created_at"`
	Output    []any          `json:"output"`
	Usage     *ResponseUsage `json:"usage,omitempty"`

	httpHeader
}

// CompactResponseRequest contains the response context to compact.
type CompactResponseRequest = CreateResponseRequest

// CreateResponse creates a non-streaming model response.
func (c *Client) CreateResponse(
	ctx context.Context,
	request CreateResponseRequest,
) (response CreateResponseResponse, err error) {
	if request.Stream {
		return response, ErrResponseStreamNotSupported
	}

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(responsesSuffix), withBody(request))
	if err != nil {
		return response, err
	}
	err = c.sendRequest(req, &response)
	return response, err
}

// RetrieveResponse gets a stored response by ID.
func (c *Client) RetrieveResponse(
	ctx context.Context,
	responseID string,
	options ...RetrieveResponseOptions,
) (response CreateResponseResponse, err error) {
	values := url.Values{}
	if len(options) > 0 {
		for _, include := range options[0].Include {
			values.Add("include", string(include))
		}
		if options[0].IncludeObfuscation != nil {
			values.Set("include_obfuscation", strconv.FormatBool(*options[0].IncludeObfuscation))
		}
		if options[0].StartingAfter != nil {
			values.Set("starting_after", strconv.Itoa(*options[0].StartingAfter))
		}
	}

	urlSuffix := responseResourceSuffix(responseID, "", values)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return response, err
	}
	err = c.sendRequest(req, &response)
	return response, err
}

// GetResponse is an alias for RetrieveResponse.
func (c *Client) GetResponse(
	ctx context.Context,
	responseID string,
	options ...RetrieveResponseOptions,
) (CreateResponseResponse, error) {
	return c.RetrieveResponse(ctx, responseID, options...)
}

// DeleteResponse deletes a stored response.
func (c *Client) DeleteResponse(ctx context.Context, responseID string) (response ResponseDeleteResponse, err error) {
	urlSuffix := responseResourceSuffix(responseID, "", nil)
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix))
	if err != nil {
		return response, err
	}
	err = c.sendRequest(req, &response)
	return response, err
}

// CancelResponse cancels a background response.
func (c *Client) CancelResponse(ctx context.Context, responseID string) (response CreateResponseResponse, err error) {
	urlSuffix := responseResourceSuffix(responseID, "cancel", nil)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix))
	if err != nil {
		return response, err
	}
	err = c.sendRequest(req, &response)
	return response, err
}

// ListResponseInputItems lists the input items for a response.
func (c *Client) ListResponseInputItems(
	ctx context.Context,
	responseID string,
	options ...ResponseInputItemsListOptions,
) (response ResponseInputItemsList, err error) {
	values := url.Values{}
	if len(options) > 0 {
		if options[0].After != "" {
			values.Set("after", options[0].After)
		}
		for _, include := range options[0].Include {
			values.Add("include", string(include))
		}
		if options[0].Limit != 0 {
			values.Set("limit", strconv.Itoa(options[0].Limit))
		}
		if options[0].Order != "" {
			values.Set("order", options[0].Order)
		}
	}

	urlSuffix := responseResourceSuffix(responseID, "input_items", values)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return response, err
	}
	err = c.sendRequest(req, &response)
	return response, err
}

// CountResponseInputTokens returns the number of input tokens a request would use.
func (c *Client) CountResponseInputTokens(
	ctx context.Context,
	request ResponseInputTokensRequest,
) (response ResponseInputTokensResponse, err error) {
	request.Stream = false
	request.StreamOptions = nil
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(responsesSuffix+"/input_tokens"),
		withBody(request),
	)
	if err != nil {
		return response, err
	}
	err = c.sendRequest(req, &response)
	return response, err
}

// CompactResponse compacts a response context for use in a later request.
func (c *Client) CompactResponse(
	ctx context.Context,
	request CompactResponseRequest,
) (response ResponseCompaction, err error) {
	request.Stream = false
	request.StreamOptions = nil
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(responsesSuffix+"/compact"),
		withBody(request),
	)
	if err != nil {
		return response, err
	}
	err = c.sendRequest(req, &response)
	return response, err
}

func responseResourceSuffix(responseID, action string, values url.Values) string {
	suffix := fmt.Sprintf("%s/%s", responsesSuffix, url.PathEscape(responseID))
	if action != "" {
		suffix += "/" + action
	}
	if len(values) != 0 {
		suffix += "?" + values.Encode()
	}
	return suffix
}
