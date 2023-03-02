package gogpt

import (
	"context"
	"errors"
	"github.com/sashabaranov/go-gpt3/internal/test"
	"io"
	"testing"
)

func TestClient_CreateChatCompletionStream(t *testing.T) {
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

	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		t.Errorf("CreateCompletionStream returned error: %v", err)
	}
	defer stream.Close()

	for {
		_, streamErr := stream.Recv()
		if errors.Is(streamErr, io.EOF) {
			return
		}

		if streamErr != nil {
			t.Errorf("stream.Recv() failed: %v", streamErr)
			return
		}
	}
}
