package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = openai.GPT6Dot1Sol
	}
	if err := run(context.Background(), openai.NewClient(key), model); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, client *openai.Client, model string) error {
	stream, err := client.CreateResponseStream(ctx, openai.CreateResponseRequest{
		Model: model, Input: "Explain server-sent events in three sentences.",
		Reasoning: &openai.ResponseReasoning{Effort: openai.ReasoningEffortLow},
	})
	if err != nil {
		return err
	}
	defer stream.Close()
	completed := false
	for {
		event, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			if !completed {
				return errors.New("stream ended before response.completed")
			}
			fmt.Println()
			return nil
		}
		if recvErr != nil {
			return recvErr
		}
		switch event.Type { //nolint:exhaustive // This text example handles text deltas and terminal events only.
		case openai.ResponseStreamEventOutputTextDelta:
			fmt.Print(event.Delta)
		case openai.ResponseStreamEventCompleted:
			completed = true
		case openai.ResponseStreamEventFailed, openai.ResponseStreamEventIncomplete, openai.ResponseStreamEventError:
			return fmt.Errorf("response did not complete: %s", event.Raw)
		}
	}
}
