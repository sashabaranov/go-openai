package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type NewsItem struct {
	Title   string
	Summary string
	Source  string
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable")
		return
	}

	client := openai.NewClient(apiKey)

	ctx := context.Background()

	req := openai.CreateResponseRequest{
		Model: openai.GPT5,
		Input: "what was a positive news story from today?",
		Tools: []openai.Tool{{
			Type: openai.ToolTypeWebSearch,
			Parameters: map[string]any{
				"filters": map[string]any{
					"allowed_domains": []string{
						"pubmed.ncbi.nlm.nih.gov",
						"clinicaltrials.gov",
						"www.who.int",
						"www.cdc.gov",
						"www.fda.gov",
					},
				},
			},
		}},
	}

	fmt.Println("Sending request to /v1/responses with web_search tool...")
	resp, err := client.CreateResponse(ctx, req)
	if err != nil {
		fmt.Printf("CreateResponse error: %v\n", err)
		return
	}
	printPositiveNews(resp)
}

// printPositiveNews extracts and prints positive news stories from the ResponsesAPIResponse.
func printPositiveNews(resp openai.CreateResponseResponse) {
	items := extractNewsItems(resp)
	printNewsItems(items)
}

func extractNewsItems(resp openai.CreateResponseResponse) []NewsItem {
	var items []NewsItem
	for _, item := range resp.Output {
		m, ok1 := item.(map[string]any)
		if !ok1 {
			continue
		}
		content, hasContent := m["content"]
		if !hasContent {
			continue
		}
		contentList, ok2 := content.([]any)
		if !ok2 {
			continue
		}
		for _, c := range contentList {
			cmap, ok3 := c.(map[string]any)
			if !ok3 {
				continue
			}
			text, _ := cmap["text"].(string)
			annotations, _ := cmap["annotations"].([]any)
			title, url := extractTitleAndURL(annotations)
			items = append(items, NewsItem{
				Title:   title,
				Summary: text,
				Source:  url,
			})
		}
	}
	return items
}

func printNewsItems(items []NewsItem) {
	if len(items) == 0 {
		fmt.Println("No results found.")
		return
	}
	fmt.Println("Positive News Stories:")
	for _, n := range items {
		fmt.Println("---")
		if n.Title != "" {
			fmt.Printf("Title: %s\n", n.Title)
		}
		if n.Summary != "" {
			fmt.Printf("Summary: %s\n", n.Summary)
		}
		if n.Source != "" {
			fmt.Printf("Source: %s\n", n.Source)
		}
	}
}

func extractTitleAndURL(annotations []any) (string, string) {
	title := ""
	url := ""
	for _, ann := range annotations {
		aMap, okAnn := ann.(map[string]any)
		if !okAnn {
			continue
		}
		if t, okTitle := aMap["title"].(string); okTitle {
			title = t
		}
		if u, okURL := aMap["url"].(string); okURL {
			url = u
		}
	}
	return title, url
}
