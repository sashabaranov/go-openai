# Go OpenAI

[![Go Reference](https://pkg.go.dev/badge/github.com/sashabaranov/go-openai.svg)](https://pkg.go.dev/github.com/sashabaranov/go-openai)
[![codecov](https://codecov.io/gh/sashabaranov/go-openai/branch/master/graph/badge.svg?token=bCbIfHLIsW)](https://codecov.io/gh/sashabaranov/go-openai)

An unofficial Go client for the [OpenAI API](https://developers.openai.com/api/docs/overview).

For new text-generation, reasoning, tool-calling, and multi-turn integrations,
start with the [Responses API](https://developers.openai.com/api/docs/guides/migrate-to-responses).
Chat Completions remains available for existing integrations.

The client also covers embeddings, images, audio, moderation, files, fine-tuning,
batches, vector stores, and legacy Assistants API surfaces.

**Building agents?** Try [Unreal Agent](https://github.com/unreallabsai/unreal-agent) - Go-based, fully async harness that drives 40% cost savings compared to Codex!

## Installation

```sh
go get github.com/sashabaranov/go-openai
```

Go OpenAI requires Go 1.18 or later.

## Quick start: Responses API

Set an [OpenAI API key](https://platform.openai.com/api-keys) in your environment:

```sh
export OPENAI_API_KEY="<your key>"
```

Then create a response and read its generated text:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	response, err := client.CreateResponse(context.Background(), openai.CreateResponseRequest{
		Model:        openai.GPT5Dot6Sol,
		Instructions: "You are a concise technical explainer.",
		Input:        "Why is the sky blue?",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.GetOutputText())
}
```

`Input` can be a string or a slice of typed input items. For reasoning, tools,
multimodal output, or custom processing, inspect `response.Output` instead of
using the `GetOutputText` convenience method.

### Continue a conversation

Use `PreviousResponseID` when OpenAI should carry the earlier response context.
Resend `Instructions` on each call when they should continue to apply.

```go
store := true

first, err := client.CreateResponse(ctx, openai.CreateResponseRequest{
	Model:        openai.GPT5Dot6Sol,
	Instructions: "Answer as a travel guide.",
	Input:        "What should I see in Lisbon?",
	Store:        &store,
})
if err != nil {
	return err
}

second, err := client.CreateResponse(ctx, openai.CreateResponseRequest{
	Model:              openai.GPT5Dot6Sol,
	Instructions:       "Answer as a travel guide.",
	Input:              "Which one is best on a rainy day?",
	PreviousResponseID: first.ID,
	Store:              &store,
})
if err != nil {
	return err
}

fmt.Println(second.GetOutputText())
```

### Stream output

```go
stream, err := client.CreateResponseStream(ctx, openai.CreateResponseRequest{
	Model: openai.GPT5Dot6Sol,
	Input: "Write a short story about a curious gopher.",
})
if err != nil {
	return err
}
defer stream.Close()

for {
	event, err := stream.Recv()
	if errors.Is(err, io.EOF) {
		break
	}
	if err != nil {
		return err
	}
	if event.Type == openai.ResponseStreamEventOutputTextDelta {
		fmt.Print(event.Delta)
	}
}
```

## Choosing a model

Choose a model based on the workload's reasoning, latency, and cost requirements.

| Constant | Model ID | Typical use |
| --- | --- | --- |
| `GPT6Dot1Sol` | `gpt-6.1-sol` | Complex coding and professional work |
| `GPT6Astra` | `gpt-6-astra` | Most demanding reasoning and coding |
| `GPT6Sol` | `gpt-6-sol` | Previous Sol model |
| `GPT6Luna` | `gpt-6-luna` | Focused, high-volume work |

GPT-5.6 and earlier model constants remain available. GPT-6.1 Sol and Astra support
`low`, `medium` (default), `high`, `xhigh`, and `max` reasoning; they do not support
`none` or `minimal`. GPT-6 Sol and Luna also support `none`. Use Responses for tool
calling with GPT-6.1 Sol or Astra, or when combining GPT-6 reasoning with tools.
See [GPT-6 guidance](https://developers.openai.com/api/docs/guides/latest-model).

See the [OpenAI model catalog](https://developers.openai.com/api/docs/models) for
capabilities and availability. Model IDs are accepted as strings, so you can use
a model before a named constant is added to this package.

## Chat Completions

Chat Completions remains supported for existing integrations:

```go
response, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
	Model: openai.GPT4oMini,
	Messages: []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Hello!",
		},
	},
})
if err != nil {
	return err
}

fmt.Println(response.Choices[0].Message.Content)
```

For a new integration, prefer Responses unless you specifically need the Chat
Completions request or response shape.

## Read batch output

Use `BatchRequestOutput[T]` to decode each line from a batch output or error file.
For example, `BatchRequestOutput[ChatCompletionResponse]` gives a typed chat body.
Use `CompletionResponse`, `EmbeddingResponse`, or `CreateResponseResponse` for the
other batch endpoints. `BatchResponse` still describes the whole batch.

A response body can also contain an API error. Use `json.RawMessage` when you need
to inspect `status_code` before decoding the body. A batch error, such as an expired
request, has a null `response` and a separate `error` with a code and message.

This example reads a chat batch file from `GetFileContent`. Pass the batch's
`OutputFileID` or `ErrorFileID` when available. Match results to inputs with
`CustomID`, because the [output order can differ from the input order](https://developers.openai.com/api/docs/guides/batch).

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	openai "github.com/sashabaranov/go-openai"
)

func printBatchOutput(ctx context.Context, client *openai.Client, fileID string) error {
	content, err := client.GetFileContent(ctx, fileID)
	if err != nil {
		return err
	}
	defer content.Close()

	decoder := json.NewDecoder(content)
	for {
		var output openai.BatchRequestOutput[json.RawMessage]
		if err := decoder.Decode(&output); errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}
		if output.Error != nil {
			fmt.Printf("%s: %s: %s\n", output.CustomID, output.Error.Code, output.Error.Message)
			continue
		}
		if output.Response == nil {
			return fmt.Errorf("missing response for %s", output.CustomID)
		}
		response := output.Response
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			var body openai.ErrorResponse
			if err := json.Unmarshal(response.Body, &body); err != nil {
				return err
			}
			fmt.Printf("%s: HTTP %d: %v\n", output.CustomID, response.StatusCode, body.Error)
			continue
		}
		var body openai.ChatCompletionResponse
		if err := json.Unmarshal(response.Body, &body); err != nil {
			return err
		}
		for _, choice := range body.Choices {
			fmt.Printf("%s: %s\n", output.CustomID, choice.Message.Content)
		}
	}
}
```

## Configuration

Use `DefaultConfig` to customize the HTTP client, base URL, organization, or
headers before constructing a client:

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
config.BaseURL = "https://your-compatible-endpoint.example/v1"
client := openai.NewClientWithConfig(config)
```

For Azure OpenAI, start with `DefaultAzureConfig` and configure the deployment
mapping or API version required by your Azure resource.

## Error handling

API failures can be inspected with `errors.As`:

```go
var apiError *openai.APIError
if errors.As(err, &apiError) {
	fmt.Printf("OpenAI error: status=%d code=%v message=%s\n",
		apiError.HTTPStatusCode, apiError.Code, apiError.Message)
}
```

## Examples

Runnable examples live in [`examples/`](examples):

- [Responses API with multi-turn state](examples/responses)
- [Chat Completions](examples/completion)
- [Chat Completions with a function tool](examples/completion-with-tool)
- [Image generation](examples/images)
- [Speech to text](examples/voice-to-text)

To run one:

```sh
go run ./examples/responses
```

## Contributing

See the [contributing guidelines](CONTRIBUTING.md) before opening a pull request.

## Thank you

Thank you to all of the project's
[contributors](https://github.com/sashabaranov/go-openai/graphs/contributors)
and sponsors, including [Carson Kahn](https://carsonkahn.com) of
[Spindle AI](https://spindleai.com).
