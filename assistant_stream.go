package openai

import (
	"context"
	"fmt"
	"net/http"
)

type AssistantThreadRunStreamResponse struct {
	ID     string       `json:"id"`
	Object string       `json:"object"`
	Delta  MessageDelta `json:"delta,omitempty"`
}

type AssistantThreadRunStream struct {
	*streamReader[AssistantThreadRunStreamResponse]
}

func (c *Client) CreateAssistantThreadRunStream(
	ctx context.Context,
	threadID string,
	request RunRequest,
) (stream *AssistantThreadRunStream, err error) {
	request.Stream = true
	urlSuffix := fmt.Sprintf("/threads/%s/runs", threadID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantV1(),
	)
	if err != nil {
		return nil, err
	}

	resp, err := sendRequestStream[AssistantThreadRunStreamResponse](c, req)
	if err != nil {
		return nil, err
	}
	stream = &AssistantThreadRunStream{
		streamReader: resp,
	}
	return
}

func (c *Client) CreateAssistantThreadRunSubmitToolOutputStream(
	ctx context.Context,
	threadID string,
	runID string,
	request SubmitToolOutputsRequest,
) (stream *AssistantThreadRunStream, err error) {
	request.Stream = true
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/submit_tool_outputs", threadID, runID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantV1(),
	)
	if err != nil {
		return
	}

	resp, err := sendRequestStream[AssistantThreadRunStreamResponse](c, req)
	if err != nil {
		return nil, err
	}
	stream = &AssistantThreadRunStream{
		streamReader: resp,
	}
	return
}
