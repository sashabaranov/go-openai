package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable")
		return
	}

	client := openai.NewClient(apiKey)

	ctx := context.Background()

	// ChatCompletionRequest with web_search tool
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "What is the latest news about OpenAI?",
			},
		},
		Tools: []openai.Tool{
			{
				Type: openai.ToolTypeWebSearch,
			},
		},
	}

	fmt.Println("Sending request with web_search tool...")
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	msg := resp.Choices[0].Message
	if msg.Content != "" {
		fmt.Printf("Response: %s\n", msg.Content)
	}

	// If the model decides to use the tool, it will return ToolCalls
	for _, toolCall := range msg.ToolCalls {
		fmt.Printf("Tool Call: ID=%s, Type=%s\n", toolCall.ID, toolCall.Type)
		if toolCall.Type == openai.ToolTypeWebSearch {
			fmt.Printf("Web Search Query: %s\n", toolCall.Function.Arguments)
		}
	}
}
