package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()
	fmt.Println("Creating new thread")
	thread, err := c.CreateThread(ctx, openai.ThreadRequest{
		Messages: []openai.ThreadMessage{{Role: openai.ThreadMessageRoleUser, Content: "i want to go home"}},
	})

	if err != nil {
		fmt.Printf("Thread error: %v\n", err)
		return
	}

	fmt.Printf("thread: %s\n", thread.ID)
	message, err := c.CreateMessage(ctx, thread.ID, openai.MessageRequest{
		Role:    "user",
		Content: "i want to go home",
	})

	if err != nil {
		fmt.Printf("Message error: %v\n", err)
		return
	}

	fmt.Printf("Message created: %v\n", message.ID)

	stream, err := c.CreateAssistantThreadRunStream(ctx, thread.ID, openai.RunRequest{
		AssistantID: os.Getenv("ASSISTANT_ID"),
		Model:       openai.GPT4TurboPreview,
	})

	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
		return
	}

	defer stream.Close()

	fmt.Printf("Stream response: ")
	/*
	   err = stream.On("thread.run.step.delta", func (resp openai.AssistantThreadRunStreamResponse, rawData []byte) {
	       fmt.Printf("run.step.delta: %s", rawData)
	   })
	   if err != nil  {
	       fmt.Printf("Stream error: %v\n", err)
	       return
	   }
	*/
	err = stream.On("thread.message.delta", func(resp openai.AssistantThreadRunStreamResponse, rawData []byte) {
		fmt.Printf("%s", rawData)
	})
	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
		return
	}
	var requiredActionRuns = []openai.Run{}
	err = stream.On("thread.run.requires_action", func(resp openai.AssistantThreadRunStreamResponse, rawData []byte) {
		run := openai.Run{}
		err := json.Unmarshal(rawData, &run)
		if err != nil {
			fmt.Printf("run unmarshal error: %v\n", err)
			return
		}
		fmt.Printf("Stream require action: %v\n", run.RequiredAction.SubmitToolOutputs.ToolCalls)
		requiredActionRuns = append(requiredActionRuns, run)
	})
	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
		return
	}
	err = stream.Wait()
	if err != io.EOF {
		fmt.Println("\nStream finished with error", err)
		return
	}
	fmt.Println("\nStream finished")

	if len(requiredActionRuns) > 0 {
		fmt.Println("Action required")
		for _, run := range requiredActionRuns {
			toolOuputs := []openai.ToolOutput{}
			for _, call := range run.RequiredAction.SubmitToolOutputs.ToolCalls {
				output := openai.ToolOutput{
					ToolCallID: call.ID,
					Output:     true,
				}
				toolOuputs = append(toolOuputs, output)
			}

			fmt.Printf("\nSubmit tool ouputs: %v\n", toolOuputs)
			stream, err := c.CreateAssistantThreadRunSubmitToolOutputStream(ctx, run.ThreadID, run.ID, openai.SubmitToolOutputsRequest{
				ToolOutputs: toolOuputs,
			})
			if err != nil {
				fmt.Printf("Stream error: %v\n", err)
				return
			}

			defer stream.Close()

			fmt.Printf("Stream response: ")
			err = stream.On("thread.message.delta", func(resp openai.AssistantThreadRunStreamResponse, rawData []byte) {
				fmt.Printf("%s", rawData)
			})
			if err != nil {
				fmt.Printf("Stream error: %v\n", err)
				return
			}
			err = stream.Wait()
			if err != io.EOF {
				fmt.Println("\nStream finished with error", err)
				return
			}
			fmt.Println("\nStream finished")
		}
	}
}
