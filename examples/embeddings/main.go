package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable")
		return
	}

	client := openai.NewClient(apiKey)

	ctx := context.Background()

	// Create an EmbeddingRequest for the user query
	queryReq := openai.EmbeddingRequest{
		Input: []string{"The food was delicious and the waiter..."},
		Model: openai.SmallEmbedding3,
	}

	// Create embeddings
	queryResp, err := client.CreateEmbeddings(ctx, queryReq)
	if err != nil {
		fmt.Printf("Embeddings error: %v\n", err)
		return
	}

	// Create an EmbeddingRequest for the target text
	targetReq := openai.EmbeddingRequest{
		Input: []string{"The restaurant had great food and service."},
		Model: openai.SmallEmbedding3,
	}

	targetResp, err := client.CreateEmbeddings(ctx, targetReq)
	if err != nil {
		fmt.Printf("Embeddings error: %v\n", err)
		return
	}

	// Calculate dot product
	dotProduct, err := queryResp.Data[0].DotProduct(&targetResp.Data[0])
	if err != nil {
		fmt.Printf("Dot product error: %v\n", err)
		return
	}

	fmt.Printf("Dot product: %f\n", dotProduct)
}
