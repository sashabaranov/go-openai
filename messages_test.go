package openai_test

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"encoding/json"
	"fmt"
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
			if r.Method == http.MethodPost {
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
}
