package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = openai.GPT6Dot1Sol
	}

	ctx := context.Background()
	client := openai.NewClient(apiKey)
	store := true

	response, err := client.CreateResponse(ctx, openai.CreateResponseRequest{
		Model:     model,
		Reasoning: &openai.ResponseReasoning{Effort: openai.ReasoningEffortLow},
		Input:     "Explain the Responses API in one sentence.",
		Store:     &store,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response.GetOutputText())

	// Continue the conversation without resending its earlier messages.
	response, err = client.CreateResponse(ctx, openai.CreateResponseRequest{
		Model:              model,
		Reasoning:          &openai.ResponseReasoning{Effort: openai.ReasoningEffortLow},
		Input:              "Now give me a concrete use case.",
		PreviousResponseID: response.ID,
		Store:              &store,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response.GetOutputText())
}
