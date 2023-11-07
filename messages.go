package openai

import (
	"context"
	"fmt"
	"net/http"
)

const (
	messagesSuffix = "/messages"
)

type Message struct {
	Id          string           `json:"id"`
	Object      string           `json:"object"`
	CreatedAt   int              `json:"created_at"`
	ThreadId    string           `json:"thread_id"`
	Role        string           `json:"role"`
	Content     []MessageContent `json:"content"`
	FileIds     []interface{}    `json:"file_ids"`
	AssistantId string           `json:"assistant_id"`
	RunId       string           `json:"run_id"`
	Metadata    struct {
	} `json:"metadata"`

	httpHeader
}

type MessageContent struct {
	Type string      `json:"type"`
	Text MessageText `json:"text"`
}
type MessageText struct {
	Value       string        `json:"value"`
	Annotations []interface{} `json:"annotations"`
}

type MessageRequest struct {
	Role     string         `json:"role"`
	Content  string         `json:"content"`
	FileIds  []string       `json:"file_ids,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// CreateMessage creates a new message.
func (c *Client) CreateMessage(ctx context.Context, threadId string, request MessageRequest) (msg Message, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s%s", threadId, messagesSuffix)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &msg)
	return
}
