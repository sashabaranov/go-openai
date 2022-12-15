# go-gpt3
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/sashabaranov/go-gpt3)
[![Go Report Card](https://goreportcard.com/badge/github.com/sashabaranov/go-gpt3)](https://goreportcard.com/report/github.com/sashabaranov/go-gpt3)


[OpenAI GPT-3](https://beta.openai.com/) API wrapper for Go

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
		Model: "ada",
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
