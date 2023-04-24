package openai

import (
	"context"
	"fmt"
)

func ExampleNewClient() {
	cli := NewClient("your-api-key")

	resp, err := cli.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		Model: GPT3Dot5Turbo,
		Messages: []ChatCompletionMessage{
			{
				Role:    ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp)
}

func ExampleNewAzureClient() {
	cli := NewAzureClient("your-api-key", "https://your Azure OpenAI Endpoint ", "your Model deployment name")

	resp, err := cli.CreateChatCompletion(
		context.Background(),
		ChatCompletionRequest{
			Model: GPT3Dot5Turbo,
			Messages: []ChatCompletionMessage{
				{
					Role:    ChatMessageRoleUser,
					Content: "Hello Azure OpenAI!",
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}
