package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

func main() {
	config := openai.DefaultConfig("sk-xxxx")
	config.BaseURL = "https://10.20.152.76:30002/v1"
	config.HTTPClient = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		},
	}
	client := openai.NewClientWithConfig(config)

	req := openai.ChatCompletionRequest{
		Model: "HengNao-v4",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "you are a helpful chatbot",
			},
		},
	}
	fmt.Println("Conversation")
	fmt.Println("---------------------")
	fmt.Print("> ")
	// s := bufio.NewScanner(os.Stdin)
	// for s.Scan() {
	req.Messages = append(req.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "你好",
	})
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		// continue
	}
	fmt.Printf("%s\n\n", resp.Choices[0].Message.Content)
	req.Messages = append(req.Messages, resp.Choices[0].Message)
	fmt.Print("> ")
	// }

	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		return
	}

	for {
		evt, err := stream.Recv()
		if err != nil {
			return
		}

		fmt.Printf("%s", evt.Choices[0].Delta.Content)
	}
}
