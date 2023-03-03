// @Author  linxianqin  2023/3/3 10:43
package gogpt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestCreateCompletionChatStream(t *testing.T) {
	//your token
	c := NewClient("token")
	ctx := context.Background()
	requestBody := CompletionTurboRequestBody{
		Model: "gpt-3.5-turbo-0301",
	}
	systemMessages := Message{
		Role:    "system",
		Content: "You are a helpful assistant.",
	}
	userMessages := Message{
		Role:    "user",
		Content: "你好",
	}
	requestBody.Messages = append(requestBody.Messages, systemMessages, userMessages)
	stream, err := c.CreateCompletionChatStream(ctx, requestBody)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stream.Close()
	var text []string
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("gpt Stream finished")
			break
		}
		if err != nil {
			fmt.Println("gpt Stream error: ", err)
			break
		}
		for _, choice := range response.Choices {
			if choice.Delta.Content != "" {
				text = append(text, choice.Delta.Content)
			}
		}
	}
	if text != nil {
		fmt.Println(strings.Join(text, ""))
	}

}
