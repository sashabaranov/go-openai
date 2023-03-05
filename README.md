# Go OpenAI
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/sashabaranov/go-openai)
[![Go Report Card](https://goreportcard.com/badge/github.com/sashabaranov/go-openai)](https://goreportcard.com/report/github.com/sashabaranov/go-openai)

> **Note**: the repository was recently renamed from `go-gpt3` to `go-openai`

This library provides Go clients for [OpenAI API](https://platform.openai.com/). We support:

* ChatGPT
* GPT-3
* DALL·E 2
* Whisper

Installation:
```
go get github.com/sashabaranov/go-openai
```


ChatGPT example usage:

```go
package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	client := openai.NewClient("your token")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	if err != nil {
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

```



Other examples:

<details>
<summary>GPT-3 completion</summary>

```go
package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	req := openai.CompletionRequest{
		Model:     openai.GPT3Ada,
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
</details>

<details>
<summary>GPT-3 streaming completion</summary>

```go
package main

import (
	"errors"
	"context"
	"fmt"
	"io"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	req := openai.CompletionRequest{
		Model:     openai.GPT3Ada,
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
</details>

<details>
<summary>Audio Speech-To-Text</summary>

```go
package main

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: "recording.mp3",
	}
	resp, err := c.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)
}
```
</details>

<details>
<summary>Configuring proxy</summary>

```go
config := openai.DefaultConfig("token")
proxyUrl, err := url.Parse("http://localhost:{port}")
if err != nil {
	panic(err)
}
transport := &http.Transport{
	Proxy: http.ProxyURL(proxyUrl),
}
config.HTTPClient = &http.Client{
	Transport: transport,
}

c := openai.NewClientWithConfig(config)
```

See also: https://pkg.go.dev/github.com/sashabaranov/go-openai#ClientConfig
</details>
