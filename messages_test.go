package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
	"net/http"
	"testing"
)

// TestMessages Tests the messages endpoint of the API using the mocked server.
func TestMessages(t *testing.T) {
	threadID := "thread_abc123"
	messageId := "msg_abc123"

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/messages/"+messageId,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				metadata := map[string]any{}
				err := json.NewDecoder(r.Body).Decode(&metadata)
				checks.NoError(t, err, "unable to decode metadata in modify message call")

				resBytes, _ := json.Marshal(
					openai.Message{
						Id:        messageId,
						Object:    "thread.message",
						CreatedAt: 1234567890,
						ThreadId:  threadID,
						Role:      "user",
						Content: []openai.MessageContent{{
							Type: "text",
							Text: openai.MessageText{
								Value:       "How does AI work?",
								Annotations: nil,
							},
						}},
						FileIds:     nil,
						AssistantId: "",
						RunId:       "",
						Metadata:    metadata,
					})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodGet:
				resBytes, _ := json.Marshal(
					openai.Message{
						Id:        messageId,
						Object:    "thread.message",
						CreatedAt: 1234567890,
						ThreadId:  threadID,
						Role:      "user",
						Content: []openai.MessageContent{{
							Type: "text",
							Text: openai.MessageText{
								Value:       "How does AI work?",
								Annotations: nil,
							},
						}},
						FileIds:     nil,
						AssistantId: "",
						RunId:       "",
						Metadata:    nil,
					})
				fmt.Fprintln(w, string(resBytes))
			default:
				t.Fatalf("unsupported messages http method: %s", r.Method)
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/messages",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.Message{
					Id:        messageId,
					Object:    "thread.message",
					CreatedAt: 1234567890,
					ThreadId:  threadID,
					Role:      "user",
					Content: []openai.MessageContent{{
						Type: "text",
						Text: openai.MessageText{
							Value:       "How does AI work?",
							Annotations: nil,
						},
					}},
					FileIds:     nil,
					AssistantId: "",
					RunId:       "",
					Metadata:    nil,
				})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.MessagesList{
					Messages: []openai.Message{{
						Id:        messageId,
						Object:    "thread.message",
						CreatedAt: 1234567890,
						ThreadId:  threadID,
						Role:      "user",
						Content: []openai.MessageContent{{
							Type: "text",
							Text: openai.MessageText{
								Value:       "How does AI work?",
								Annotations: nil,
							},
						}},
						FileIds:     nil,
						AssistantId: "",
						RunId:       "",
						Metadata:    nil,
					}}})
				fmt.Fprintln(w, string(resBytes))
			default:
				t.Fatalf("unsupported messages http method: %s", r.Method)
			}
		},
	)

	ctx := context.Background()

	// static assertion of return type
	var msg openai.Message
	msg, err := client.CreateMessage(ctx, threadID, openai.MessageRequest{
		Role:     "user",
		Content:  "How does AI work?",
		FileIds:  nil,
		Metadata: nil,
	})
	checks.NoError(t, err, "CreateMessage error")

	var msgs openai.MessagesList
	msgs, err = client.ListMessage(ctx, threadID, nil, nil, nil, nil)
	checks.NoError(t, err, "ListMessages error")
	if len(msgs.Messages) != 1 {
		t.Fatalf("unexpected length of fetched messages")
	}

	msg, err = client.RetrieveMessage(ctx, threadID, messageId)
	checks.NoError(t, err, "RetrieveMessage error")
	if msg.Id != messageId {
		t.Fatalf("unexpected message id: '%s'", msg.Id)
	}

	msg, err = client.ModifyMessage(ctx, threadID, messageId,
		map[string]any{
			"foo": "bar",
		})
	checks.NoError(t, err, "ModifyMessage error")
	if msg.Metadata["foo"] != "bar" {
		t.Fatalf("expected message metadata to get modified")
	}
}
