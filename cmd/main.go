//nolint:forbidigo // test for a possible user consuming this library
package main

import (
	"context"
	"fmt"
	"os"

	gogpt "github.com/sashabaranov/go-gpt3"
)

func main() {
	c := gogpt.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	request := gogpt.CompletionRequest{
		Prompt: "Ex falso quodlibet",
		Model:  "text-davinci-002",
		//nolint:gomnd // this is test for a user consuming this library
		MaxTokens: 20,
		Stream:    true,
	}

	fmt.Println("starting stream:")

	responses, err := c.CreateCompletionStream(ctx, request)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, resp := range responses {
		fmt.Println(resp.Choices[0].Text)
	}
}
