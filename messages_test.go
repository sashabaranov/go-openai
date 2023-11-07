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

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/messages",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.Message{
					Id:        "msg_abc123",
					Object:    "thread.message",
					CreatedAt: 1234567890,
					ThreadId:  "thread_abc123",
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
					Metadata:    struct{}{},
				})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.MessagesList{
					Messages: []openai.Message{{
						Id:        "msg_abc123",
						Object:    "thread.message",
						CreatedAt: 1234567890,
						ThreadId:  "thread_abc123",
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
						Metadata:    struct{}{},
					}}})
				fmt.Fprintln(w, string(resBytes))
			default:
				t.Fatalf("unsupported messages http method: %s", r.Method)
			}
		},
	)

	ctx := context.Background()

	_, err := client.CreateMessage(ctx, threadID, openai.MessageRequest{
		Role:     "user",
		Content:  "How does AI work?",
		FileIds:  nil,
		Metadata: nil,
	})
	checks.NoError(t, err, "CreateMessage error")

	_, err = client.ListMessage(ctx, threadID, nil, nil, nil, nil)
	checks.NoError(t, err, "ListMessages error")
}
