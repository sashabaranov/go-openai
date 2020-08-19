# go-gpt3
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/sashabaranov/go-gpt3)

[OpenAI GPT-3](https://beta.openai.com/) API for Go


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
		MaxTokens: 5,
		Prompt:    "Lorem ipsum",
	}
	resp, err := c.CreateCompletion(ctx, "ada", req)
	if err != nil {
		return
	}
	fmt.Println(resp.Сhoices[0].Text)

	searchReq := gogpt.SearchRequest{
		Documents: []string{"White House", "hospital", "school"},
		Query:     "the president",
	}
	searchResp, err := c.Search(ctx, "ada", searchReq)
	if err != nil {
		return
	}
	fmt.Println(searchResp.SearchResults)
}
```
