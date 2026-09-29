package openai

import (
	"context"
	"encoding/json"
	"net/http"
)

// ResponseStreamEventType identifies an event emitted while a response is generated.
type ResponseStreamEventType string

const (
	ResponseStreamEventCreated                     ResponseStreamEventType = "response.created"
	ResponseStreamEventQueued                      ResponseStreamEventType = "response.queued"
	ResponseStreamEventInProgress                  ResponseStreamEventType = "response.in_progress"
	ResponseStreamEventCompleted                   ResponseStreamEventType = "response.completed"
	ResponseStreamEventFailed                      ResponseStreamEventType = "response.failed"
	ResponseStreamEventIncomplete                  ResponseStreamEventType = "response.incomplete"
	ResponseStreamEventOutputItemAdded             ResponseStreamEventType = "response.output_item.added"
	ResponseStreamEventOutputItemDone              ResponseStreamEventType = "response.output_item.done"
	ResponseStreamEventContentPartAdded            ResponseStreamEventType = "response.content_part.added"
	ResponseStreamEventContentPartDone             ResponseStreamEventType = "response.content_part.done"
	ResponseStreamEventOutputTextDelta             ResponseStreamEventType = "response.output_text.delta"
	ResponseStreamEventOutputTextDone              ResponseStreamEventType = "response.output_text.done"
	ResponseStreamEventOutputTextAnnotationAdded   ResponseStreamEventType = "response.output_text.annotation.added"
	ResponseStreamEventRefusalDelta                ResponseStreamEventType = "response.refusal.delta"
	ResponseStreamEventRefusalDone                 ResponseStreamEventType = "response.refusal.done"
	ResponseStreamEventFunctionArgumentsDelta      ResponseStreamEventType = "response.function_call_arguments.delta"
	ResponseStreamEventFunctionArgumentsDone       ResponseStreamEventType = "response.function_call_arguments.done"
	ResponseStreamEventReasoningSummaryTextDelta   ResponseStreamEventType = "response.reasoning_summary_text.delta"
	ResponseStreamEventReasoningSummaryTextDone    ResponseStreamEventType = "response.reasoning_summary_text.done"
	ResponseStreamEventReasoningSummaryPartAdded   ResponseStreamEventType = "response.reasoning_summary_part.added"
	ResponseStreamEventReasoningSummaryPartDone    ResponseStreamEventType = "response.reasoning_summary_part.done"
	ResponseStreamEventReasoningTextDelta          ResponseStreamEventType = "response.reasoning_text.delta"
	ResponseStreamEventReasoningTextDone           ResponseStreamEventType = "response.reasoning_text.done"
	ResponseStreamEventAudioDelta                  ResponseStreamEventType = "response.audio.delta"
	ResponseStreamEventAudioDone                   ResponseStreamEventType = "response.audio.done"
	ResponseStreamEventAudioTranscriptDelta        ResponseStreamEventType = "response.audio.transcript.delta"
	ResponseStreamEventAudioTranscriptDone         ResponseStreamEventType = "response.audio.transcript.done"
	ResponseStreamEventWebSearchInProgress         ResponseStreamEventType = "response.web_search_call.in_progress"
	ResponseStreamEventWebSearchSearching          ResponseStreamEventType = "response.web_search_call.searching"
	ResponseStreamEventWebSearchCompleted          ResponseStreamEventType = "response.web_search_call.completed"
	ResponseStreamEventFileSearchInProgress        ResponseStreamEventType = "response.file_search_call.in_progress"
	ResponseStreamEventFileSearchSearching         ResponseStreamEventType = "response.file_search_call.searching"
	ResponseStreamEventFileSearchCompleted         ResponseStreamEventType = "response.file_search_call.completed"
	ResponseStreamEventCodeInterpreterInProgress   ResponseStreamEventType = "response.code_interpreter_call.in_progress"
	ResponseStreamEventCodeInterpreterInterpreting ResponseStreamEventType = "response.code_interpreter_call.interpreting"
	ResponseStreamEventCodeInterpreterCompleted    ResponseStreamEventType = "response.code_interpreter_call.completed"
	ResponseStreamEventCodeInterpreterCodeDelta    ResponseStreamEventType = "response.code_interpreter_call_code.delta"
	ResponseStreamEventCodeInterpreterCodeDone     ResponseStreamEventType = "response.code_interpreter_call_code.done"
	ResponseStreamEventCustomToolInputDelta        ResponseStreamEventType = "response.custom_tool_call_input.delta"
	ResponseStreamEventCustomToolInputDone         ResponseStreamEventType = "response.custom_tool_call_input.done"
	ResponseStreamEventImageGenerationInProgress   ResponseStreamEventType = "response.image_generation_call.in_progress"
	ResponseStreamEventImageGenerationGenerating   ResponseStreamEventType = "response.image_generation_call.generating"
	ResponseStreamEventImageGenerationCompleted    ResponseStreamEventType = "response.image_generation_call.completed"
	ResponseStreamEventImageGenerationPartialImage ResponseStreamEventType = "response.image_generation_call.partial_image"
	ResponseStreamEventMCPCallInProgress           ResponseStreamEventType = "response.mcp_call.in_progress"
	ResponseStreamEventMCPCallCompleted            ResponseStreamEventType = "response.mcp_call.completed"
	ResponseStreamEventMCPCallFailed               ResponseStreamEventType = "response.mcp_call.failed"
	ResponseStreamEventMCPCallArgumentsDelta       ResponseStreamEventType = "response.mcp_call_arguments.delta"
	ResponseStreamEventMCPCallArgumentsDone        ResponseStreamEventType = "response.mcp_call_arguments.done"
	ResponseStreamEventMCPListToolsInProgress      ResponseStreamEventType = "response.mcp_list_tools.in_progress"
	ResponseStreamEventMCPListToolsCompleted       ResponseStreamEventType = "response.mcp_list_tools.completed"
	ResponseStreamEventMCPListToolsFailed          ResponseStreamEventType = "response.mcp_list_tools.failed"
	ResponseStreamEventError                       ResponseStreamEventType = "error"
)

// ResponseStreamEvent contains the common fields across Responses API SSE event variants.
type ResponseStreamEvent struct {
	Type              ResponseStreamEventType `json:"type"`
	SequenceNumber    int                     `json:"sequence_number,omitempty"`
	Response          *CreateResponseResponse `json:"response,omitempty"`
	Item              *ResponseOutputItem     `json:"item,omitempty"`
	Part              *ResponseOutputContent  `json:"part,omitempty"`
	Annotation        *ResponseAnnotation     `json:"annotation,omitempty"`
	ItemID            string                  `json:"item_id,omitempty"`
	OutputIndex       int                     `json:"output_index,omitempty"`
	ContentIndex      int                     `json:"content_index,omitempty"`
	SummaryIndex      int                     `json:"summary_index,omitempty"`
	Delta             string                  `json:"delta,omitempty"`
	Text              string                  `json:"text,omitempty"`
	Arguments         string                  `json:"arguments,omitempty"`
	PartialImageB64   string                  `json:"partial_image_b64,omitempty"`
	PartialImageIndex int                     `json:"partial_image_index,omitempty"`
	Logprobs          []ResponseLogprob       `json:"logprobs,omitempty"`
	Code              string                  `json:"code,omitempty"`
	Message           string                  `json:"message,omitempty"`
	Param             any                     `json:"param,omitempty"`
	Error             *ResponseError          `json:"error,omitempty"`
	Obfuscation       string                  `json:"obfuscation,omitempty"`
	Raw               json.RawMessage         `json:"-"`
}

// UnmarshalJSON decodes known event fields and retains the complete event for
// forward compatibility with event variants added by the API.
func (e *ResponseStreamEvent) UnmarshalJSON(data []byte) error {
	type eventAlias ResponseStreamEvent
	if err := json.Unmarshal(data, (*eventAlias)(e)); err != nil {
		return err
	}
	e.Raw = append(e.Raw[:0], data...)
	return nil
}

// ResponseStream reads server-sent events from a streaming Responses API request.
type ResponseStream struct {
	*streamReader[ResponseStreamEvent]
}

// CreateResponseStream creates a response and streams its generation events.
func (c *Client) CreateResponseStream(
	ctx context.Context,
	request CreateResponseRequest,
) (stream *ResponseStream, err error) {
	request.Stream = true
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(responsesSuffix),
		withBody(request),
	)
	if err != nil {
		return nil, err
	}

	reader, err := sendRequestStream[ResponseStreamEvent](c, req)
	if err != nil {
		return nil, err
	}
	return &ResponseStream{streamReader: reader}, nil
}
