package main

import (
	"context"
	"fmt"
	"log"
	"os"

	. "github.com/sashabaranov/go-gpt3"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatalln("Missing API KEY")
	}

	client := NewClient(apiKey)

	fmt.Print("\n\nstarting stream:\n")

	request := CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     "text-davinci-002",
		MaxTokens: 20,
		Stream:    true,
	}

	ctx := context.Background()
	responses, err := client.CreateCompletionStream(ctx, request)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, response := range responses {
		fmt.Println(response.Choices[0].Text)
	}
}
