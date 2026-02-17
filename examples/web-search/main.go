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

	req := openai.CreateResponseRequest{
		Model: "gpt-5",
		Input: "what was a positive news story from today?",
		Tools: []openai.Tool{{Type: openai.ToolTypeWebSearch}},
	}

	fmt.Println("Sending request to /v1/responses with web_search tool...")
	resp, err := client.CreateResponse(ctx, req)
	if err != nil {
		fmt.Printf("CreateResponse error: %v\n", err)
		return
	}
	// Extract and print positive news stories from the response
	if len(resp.Output) == 0 {
		fmt.Println("No results found.")
		return
	}
	fmt.Println("Positive News Stories:")
	for _, item := range resp.Output {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		content, hasContent := m["content"]
		if !hasContent {
			continue
		}
		contentList, ok := content.([]any)
		if !ok {
			continue
		}
		for _, c := range contentList {
			cmap, ok := c.(map[string]any)
			if !ok {
				continue
			}
			text, _ := cmap["text"].(string)
			annotations, _ := cmap["annotations"].([]any)
			url := ""
			title := ""
			for _, a := range annotations {
				aMap, ok := a.(map[string]any)
				if !ok {
					continue
				}
				if t, ok := aMap["title"].(string); ok {
					title = t
				}
				if u, ok := aMap["url"].(string); ok {
					url = u
				}
			}
			fmt.Println("---")
			if title != "" {
				fmt.Printf("Title: %s\n", title)
			}
			if text != "" {
				fmt.Printf("Summary: %s\n", text)
			}
			if url != "" {
				fmt.Printf("Source: %s\n", url)
			}
		}
	}
}
