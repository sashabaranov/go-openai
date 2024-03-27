package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"io"
	"os"
)

func main() {
	c := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 200,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "hi👋",
			},
		},
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	/*
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				fmt.Println("\nStream finished")
				return
			}

			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				return
			}

			fmt.Printf("%s", response.Choices[0].Delta.Content)
		}
	*/
	err = stream.On("message", func(resp openai.ChatCompletionStreamResponse, rawData []byte) {
		fmt.Printf("%s", resp.Choices[0].Delta.Content)
	})
	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
		return
	}
	err = stream.Wait()
	if err != io.EOF {
		fmt.Println("\nStream finished with error", err)
		return
	}
	fmt.Println("\nStream finished")
}
