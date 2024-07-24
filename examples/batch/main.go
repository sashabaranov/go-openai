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
	response, err := createBatch(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("batchID: %+v\n", response.ID)

	// retrieve Batch
	//batchID := "batch_XXXXXXXXXXXXX"
	//retrieveBatch(ctx, client, batchID)
}

func createBatch(ctx context.Context, client *openai.Client) (openai.BatchResponse, error) {
	req := openai.CreateBatchWithUploadFileRequest{
		Endpoint: openai.BatchEndpointChatCompletions,
	}
	comments := []string{
		"it's a good bike but if you have a problem after the sale they either do not respond to you or the parts are not available",
		"I ordered 2 Mars 2.0.A blue and an Orange.Blue came first and had shipping damage to the seat post.It came with a flip seat.The Orange came  about 10 days later and didnt have a flip seat.I notified customer service about both issues.They shipped a new seat post but it will not fit the blue bike because it is for a non flip seat.I am still waiting for a fix both both of these problems.\nI do not like the fact that the throttle cannot be used without the peddle assist being on.At time I feel the peddle assist is dangerous.You better not try to make a turn with the peddle assist on.",
		"This was my first E-bike. Love it so far, it has plenty power and range. I use it for hunting on our land. Works well for me, I am very satisfied.",
		"I would definitely recommend this bike. Easy to use. Great battery life, quick delivery!",
		"Slight difficulty setting up bike but it’s perfect and love it’s speed and power",
	}
	prompt := "Please analyze the following product review and extract the mentioned dimensions and reasons.\n\nReview example:\n```\nThese headphones have excellent sound quality, perfect for music lovers. I wear them every day during my commute, and the noise cancellation is great. The customer service is also very good; they patiently solved my issues. The only downside is that wearing them for long periods makes my ears hurt.\n```\n\nExpected JSON output example:\n```json\n{\n    \"dimensions\": [\n        {\n            \"dimension\": \"Usage Scenario\",\n            \"value\": \"during commute\",\n            \"reason\": \"user wears them every day during commute\"\n        },\n        {\n            \"dimension\": \"Target Audience\",\n            \"value\": \"music lovers\",\n            \"reason\": \"user is a music lover\"\n        },\n        {\n            \"dimension\": \"Positive Experience\",\n            \"value\": \"excellent sound quality\",\n            \"reason\": \"user thinks the headphones have excellent sound quality\"\n        },\n        {\n            \"dimension\": \"Positive Experience\",\n            \"value\": \"great noise cancellation\",\n            \"reason\": \"user thinks the noise cancellation is great\"\n        },\n        {\n            \"dimension\": \"Negative Experience\",\n            \"value\": \"ears hurt after long periods\",\n            \"reason\": \"user thinks wearing them for long periods makes ears hurt\"\n        }\n    ]\n}\n```\nPlease analyze accordingly and return the results in JSON format."

	for i, comment := range comments {
		req.AddChatCompletion(fmt.Sprintf("req-%d", i), openai.ChatCompletionRequest{
			Model: openai.GPT4oMini20240718,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: prompt},
				{Role: openai.ChatMessageRoleUser, Content: comment},
			},
			MaxTokens: 2000,
		})
	}
	return client.CreateBatchWithUploadFile(ctx, req)
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
