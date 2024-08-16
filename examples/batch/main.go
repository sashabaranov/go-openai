package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"os"
)

func main() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	// create batch
	response, err := createBatchChatCompletion(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("batchID: %s\n", response.ID)

	// retrieve Batch
	//batchID := "batch_XXXXXXXXXXXXX"
	//retrieveBatch(ctx, client, batchID)
}

func createBatchChatCompletion(ctx context.Context, client *openai.Client) (openai.BatchResponse, error) {
	var chatCompletions = make([]openai.BatchChatCompletion, 5)
	for i := 0; i < 5; i++ {
		chatCompletions[i] = openai.BatchChatCompletion{
			CustomID: fmt.Sprintf("req-%d", i),
			ChatCompletion: openai.ChatCompletionRequest{
				Model: openai.GPT4oMini,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("What is the square of %d?", i+1),
					},
				},
			},
		}
	}

	return client.CreateBatchWithChatCompletions(ctx, openai.CreateBatchWithChatCompletionsRequest{
		ChatCompletions: chatCompletions,
	})
}

func retrieveBatch(ctx context.Context, client *openai.Client, batchID string) {
	batch, err := client.RetrieveBatch(ctx, batchID)
	if err != nil {
		return
	}
	fmt.Printf("batchStatus: %s\n", batch.Status)

	files := map[string]*string{
		"inputFile":  &batch.InputFileID,
		"outputFile": batch.OutputFileID,
		"errorFile":  batch.ErrorFileID,
	}
	for name, fileID := range files {
		if fileID != nil {
			content, err := client.GetFileContent(ctx, *fileID)
			if err != nil {
				return
			}
			all, _ := io.ReadAll(content)
			fmt.Printf("%s: %s\n", name, all)
		}
	}
}
