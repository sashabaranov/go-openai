package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}
	client := openai.NewClient(key)
	response, err := client.CreateResponse(context.Background(), openai.CreateResponseRequest{
		Model: openai.GPT6Dot1Sol,
		Input: "Compare a monolith and microservices for a small engineering team. " +
			"Delegate reliability and development costs to two subagents, then synthesize their findings.",
		Reasoning:  &openai.ResponseReasoning{Effort: openai.ReasoningEffortLow},
		Betas:      []string{openai.ResponseBetaMultiAgent},
		MultiAgent: &openai.ResponseMultiAgent{Enabled: true, MaxConcurrentSubagents: 2},
	})
	if err != nil {
		log.Fatal(err)
	}
	if response.Status != openai.ResponseStatusCompleted {
		log.Fatalf("response status %s: error=%+v incomplete=%+v",
			response.Status, response.Error, response.IncompleteDetails)
	}
	// Output also contains subagent messages and hosted collaboration actions.
	// Only the root agent's final answer is intended for display here.
	fmt.Println(response.GetFinalOutputText())
}
