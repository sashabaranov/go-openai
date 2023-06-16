# Go OpenAI
[![Go Reference](https://pkg.go.dev/badge/github.com/sashabaranov/go-openai.svg)](https://pkg.go.dev/github.com/sashabaranov/go-openai)
[![Go Report Card](https://goreportcard.com/badge/github.com/sashabaranov/go-openai)](https://goreportcard.com/report/github.com/sashabaranov/go-openai)
[![codecov](https://codecov.io/gh/sashabaranov/go-openai/branch/master/graph/badge.svg?token=bCbIfHLIsW)](https://codecov.io/gh/sashabaranov/go-openai)

This library provides unofficial Go clients for [OpenAI API](https://platform.openai.com/). We support: 

* ChatGPT
* GPT-3, GPT-4
* DALL·E 2
* Whisper

### Installation:
```
go get github.com/sashabaranov/go-openai
```


### ChatGPT example usage:

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
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

```

### Other examples:

<details>
<summary>ChatGPT streaming completion</summary>

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 20,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Lorem ipsum",
			},
		},
		Stream: true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
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

		fmt.Printf(response.Choices[0].Delta.Content)
	}
}
```
</details>

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
		fmt.Printf("Completion error: %v\n", err)
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
		fmt.Printf("CompletionStream error: %v\n", err)
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
<summary>Audio Captions</summary>

```go
package main

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient(os.Getenv("OPENAI_KEY"))

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: os.Args[1],
		Format:   openai.AudioResponseFormatSRT,
	}
	resp, err := c.CreateTranscription(context.Background(), req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	f, err := os.Create(os.Args[1] + ".srt")
	if err != nil {
		fmt.Printf("Could not open file: %v\n", err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(resp.Text); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}
```
</details>

<details>
<summary>DALL-E 2 image generation</summary>

```go
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"image/png"
	"os"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	// Sample image by link
	reqUrl := openai.ImageRequest{
		Prompt:         "Parrot on a skateboard performs a trick, cartoon style, natural light, high detail",
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	respUrl, err := c.CreateImage(ctx, reqUrl)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}
	fmt.Println(respUrl.Data[0].URL)

	// Example image as base64
	reqBase64 := openai.ImageRequest{
		Prompt:         "Portrait of a humanoid parrot in a classic costume, high detail, realistic light, unreal engine",
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := c.CreateImage(ctx, reqBase64)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		return
	}

	r := bytes.NewReader(imgBytes)
	imgData, err := png.Decode(r)
	if err != nil {
		fmt.Printf("PNG decode error: %v\n", err)
		return
	}

	file, err := os.Create("example.png")
	if err != nil {
		fmt.Printf("File creation error: %v\n", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, imgData); err != nil {
		fmt.Printf("PNG encode error: %v\n", err)
		return
	}

	fmt.Println("The image was saved as example.png")
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

<details>
<summary>ChatGPT support context</summary>

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func main() {
	client := openai.NewClient("your token")
	messages := make([]openai.ChatCompletionMessage, 0)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			continue
		}

		content := resp.Choices[0].Message.Content
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})
		fmt.Println(content)
	}
}
```
</details>

<details>
<summary>Azure OpenAI ChatGPT</summary>

```go
package main

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	config := openai.DefaultAzureConfig("your Azure OpenAI Key", "https://your Azure OpenAI Endpoint")
	// If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
	// config.AzureModelMapperFunc = func(model string) string {
	// 	azureModelMapping = map[string]string{
	// 		"gpt-3.5-turbo": "your gpt-3.5-turbo deployment name",
	// 	}
	// 	return azureModelMapping[model]
	// }

	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello Azure OpenAI!",
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

```
</details>

<details>
<summary>Azure OpenAI Embeddings</summary>

```go
package main

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func main() {

	config := openai.DefaultAzureConfig("your Azure OpenAI Key", "https://your Azure OpenAI Endpoint")
	config.APIVersion = "2023-05-15" // optional update to latest API version

	//If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
	//config.AzureModelMapperFunc = func(model string) string {
	//    azureModelMapping = map[string]string{
	//        "gpt-3.5-turbo":"your gpt-3.5-turbo deployment name",
	//    }
	//    return azureModelMapping[model]
	//}

	input := "Text to vectorize"

	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{input},
			Model: openai.AdaEmbeddingV2,
		})

	if err != nil {
		fmt.Printf("CreateEmbeddings error: %v\n", err)
		return
	}

	vectors := resp.Data[0].Embedding // []float32 with 1536 dimensions

	fmt.Println(vectors[:10], "...", vectors[len(vectors)-10:])
}
```
</details>


<details>
<summary>Generate Embeddings</summary>

```go
package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
)

func getEmbedding(ctx context.Context, client *openai.Client, input []string) ([]float32, error) {

	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: input,
		Model: openai.AdaEmbeddingV2,
	})

	if err != nil {
		return nil, err
	}

	return resp.Data[0].Embedding, nil
}

func main() {

	ctx := context.Background()
	client := openai.NewClient("your token")

	// example selections
	selections := []string{
		"Welcome to the go-openai interface, which will be the gateway for golang software engineers to enter the OpenAI development world.",
		"It was tasty and fresh. The other one I bought was old and tasted moldy. But this one was good.",
		"Great coffee at a good price. I'm a subscription buyer and I buy this month after month. What more can I say?",
		"This chocolate is amazing..I love the taste and smell, this is the only chocolate for me...I found a new love!",
		"I love this coffee!  And such a great price.  Will buy more when I am running out which will be soon.",
		"The Raspberry Tea Syrup is great. I can use it for hot and cold drinks as well in certain recipes.",
		"Everyone that dips with this loves it!  So easy to use! Olive oil and tasty bread is all you need.",
		"This is a favorite of mine for using over ice. Even bought it to give out as Christmas gifts last year.",
		"If you like a great , hot, sauce then buy this. If spicy with heat isn't to your liking then don't buy it.",
		"My name is Aceld, and I am a Golang software development engineer. I like young and beautiful girls.",
		"The competition was held over two days,24 July and 2 August. The qualifying round was the first day with the apparatus final on the second day.",
		"There are 4 types of gymnastics apparatus: floor, vault, pommel horse, and rings. The apparatus final is a competition between the top 8 gymnasts in each apparatus.",
	}

	// Generate embeddings
	var selectionsEmbeddings [][]float32
	for _, selection := range selections {
		embedding, err := getEmbedding(ctx, client, []string{selection})
		if err != nil {
			fmt.Printf("GetEmedding error: %v\n", err)
			return
		}
		selectionsEmbeddings = append(selectionsEmbeddings, embedding)
	}

	// Write embeddings binary data to file
	file, err := os.Create("embeddings.bin")
	if err != nil {
		fmt.Printf("Create file error: %v\n", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(selectionsEmbeddings)
	if err != nil {
		fmt.Printf("Encode error: %v\n", err)
		return
	}

	return
}
```
</details>

<details>
<summary>Embedding Similarity Search</summary>

```go
package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

func getEmbedding(ctx context.Context, client *openai.Client, input []string) ([]float32, error) {
	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: input, 
		Model: openai.AdaEmbeddingV2,
	})
	
	if err != nil {
		return nil, err
	}
	
	return resp.Data[0].Embedding, nil
}

// Sort the index in descending order of similarity
func sortIndexes(scores []float32) []int {
	indexes := make([]int, len(scores))
	for i := range indexes {
		indexes[i] = i
	}
	sort.SliceStable(indexes, func(i, j int) bool {
		return scores[indexes[i]] > scores[indexes[j]]
	})
	return indexes
}

func main() {
	ctx := context.Background()
	client := openai.NewClient("your token")

	//  "embeddings.bin" from exp: <Generate Embeddings>
	file, err := os.Open("embeddings.bin")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// load all embeddings from local binary file
	var allEmbeddings [][]float32
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&allEmbeddings); err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}

	// make some input you like
	input := "I am a Golang Software Engineer, I like Go and OpenAI."

	// get embedding of input
	inputEmbd, err := getEmbedding(ctx, client, []string{input})
	if err != nil {
		fmt.Printf("GetEmedding error: %v\n", err)
		return
	}

	// Calculate similarity through cosine matching algorithm
	var questionScores []float32
	for _, embed := range allEmbeddings {
		// OpenAI embeddings are normalized to length 1, which means that:
		// Cosine similarity can be computed slightly faster using just a dot product
		score := openai.DotProduct(embed, inputEmbd)
		questionScores = append(questionScores, score)
	}

	// Take the subscripts of the top few selections with the highest similarity
	sortedIndexes := sortIndexes(questionScores)
	sortedIndexes = sortedIndexes[:3] // Top 3

	fmt.Println("input:", input)
	fmt.Println("----------------------")
	fmt.Println("similarity section:")
	selectionsFile, err := os.Open("selections.txt")
	if err != nil {
		fmt.Printf("Open file error: %v\n", err)
		return
	}
	defer selectionsFile.Close()

	fileData, err := ioutil.ReadAll(selectionsFile)
	if err != nil {
		fmt.Printf("ReadAll file error: %v\n", err)
		return
	}

	// Split by line
	selections := strings.Split(string(fileData), "\n")
	
	for _, index := range sortedIndexes {
		selection := selections[index]
		fmt.Printf("%.4f %s\n", questionScores[index], selection)
	}

	return
}
```
</details>

<details>
<summary>Error handling</summary>

Open-AI maintains clear documentation on how to [handle API errors](https://platform.openai.com/docs/guides/error-codes/api-errors)

example:
```go
e := &openai.APIError{}

if errors.As(err, &e) {
	switch e.HTTPStatusCode {
	case 401:
        // invalid auth or key (do not retry)
    case 429:
        // rate limiting or engine overload (wait and retry) 
    case 500:
        // openai server error (retry)
    default:
        // unhandled
    }
}

```
</details>

See the `examples/` folder for more.
