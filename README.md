# go-gpt3
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/sashabaranov/go-gpt3)
[![Go Report Card](https://goreportcard.com/badge/github.com/sashabaranov/go-gpt3)](https://goreportcard.com/report/github.com/sashabaranov/go-gpt3)


[OpenAI ChatGPT and GPT-3](https://platform.openai.com/) API client for Go

Installation:
```
go get github.com/sashabaranov/go-gpt3
```


Example usage:

```go
package main

import (
	"context"
	"fmt"
	gogpt "github.com/sashabaranov/go-gpt3"
)

func main() {
	c := gogpt.NewClient("your token")
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:     gogpt.GPT3Ada,
		MaxTokens: 5,
		Prompt:    "Lorem ipsum",
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		return
	}
	fmt.Println(resp.Choices[0].Text)
}
```

Streaming response example:

```go
package main

import (
	"errors"
	"context"
	"fmt"
	"io"
	gogpt "github.com/sashabaranov/go-gpt3"
)

func main() {
	c := gogpt.NewClient("your token")
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:     gogpt.GPT3Ada,
		MaxTokens: 5,
		Prompt:    "Lorem ipsum",
		Stream:    true,
	}
	stream, err := c.CreateCompletionStream(ctx, req)
	if err != nil {
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			return
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}


		fmt.Printf("Stream response: %v\n", response)
	}
}
```
