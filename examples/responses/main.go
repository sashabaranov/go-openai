package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
)

func main() {
	docs := flag.Bool("docs", false, "answer a public Go documentation question using a function callback")
	flag.Parse()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = openai.GPT5Dot6Sol
	}

	ctx := context.Background()
	client := openai.NewClient(apiKey)
	if *docs {
		docsCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
		lookup := func(ctx context.Context, query string) (docSource, error) {
			return parallelDocs(ctx, &http.Client{Timeout: 30 * time.Second}, parallelMCPURL, query)
		}
		answer, err := answerWithDocs(docsCtx, client, model, lookup)
		cancel()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(answer)
		return
	}
	store := true

	response, err := client.CreateResponse(ctx, openai.CreateResponseRequest{
		Model: model,
		Input: "Explain the Responses API in one sentence.",
		Store: &store,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response.GetOutputText())

	// Continue the conversation without resending its earlier messages.
	response, err = client.CreateResponse(ctx, openai.CreateResponseRequest{
		Model:              model,
		Input:              "Now give me a concrete use case.",
		PreviousResponseID: response.ID,
		Store:              &store,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response.GetOutputText())
}
