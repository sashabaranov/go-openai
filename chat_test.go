package gogpt

import (
	"context"
	"github.com/sashabaranov/go-gpt3/internal/test"
	"testing"
)

func TestClient_CreateChatCompletion(t *testing.T) {
	config := DefaultConfig(test.GetTestToken())
	client := NewClientWithConfig(config)
	ctx := context.Background()

	chatMessages := make([]ChatCompletionMessage, 0)
	chatMessages = append(chatMessages, ChatCompletionMessage{
		Role:    "user",
		Content: "Hello",
	})
	request := ChatCompletionRequest{
		Model:    GPT3Dot5Turbo,
		Messages: chatMessages,
	}

	_, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		t.Errorf("CreateCompletionStream returned error: %v", err)
	}
}
