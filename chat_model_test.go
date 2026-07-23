package openai_test

import (
	"context"
	"errors"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestCreateChatCompletionEmptyModel(t *testing.T) {
	c := openai.NewClient("test-key")
	_, err := c.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}},
	})
	if !errors.Is(err, openai.ErrChatCompletionRequestEmptyModel) {
		t.Fatalf("got %v", err)
	}
}
