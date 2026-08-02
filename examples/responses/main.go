package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const maxToolRounds = 5

type settings struct {
	model            string
	runBackground    bool
	runCompaction    bool
	keepResponses    bool
	requestTimeout   time.Duration
	backgroundPrompt string
}

type weatherArguments struct {
	Location string `json:"location"`
	Unit     string `json:"unit"`
}

type weatherResult struct {
	Location    string `json:"location"`
	Temperature int    `json:"temperature"`
	Unit        string `json:"unit"`
	Conditions  string `json:"conditions"`
	Source      string `json:"source"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "responses example: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	config := parseFlags()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return errors.New("OPENAI_API_KEY is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.requestTimeout)
	defer cancel()

	client := openai.NewClient(apiKey)
	request := initialRequest(config.model)

	count, err := client.CountResponseInputTokens(ctx, openai.ResponseInputTokensRequest{
		Model:        request.Model,
		Input:        request.Input,
		Instructions: request.Instructions,
		Tools:        request.Tools,
		ToolChoice:   request.ToolChoice,
		Text:         request.Text,
	})
	if err != nil {
		return fmt.Errorf("count input tokens: %w", err)
	}
	fmt.Printf("Model: %s\nInput tokens before generation: %d\n", config.model, count.InputTokens)

	var storedResponseIDs []string
	defer func() {
		if !config.keepResponses {
			deleteStoredResponses(client, storedResponseIDs)
		}
	}()

	response, responseIDs, err := createWithToolLoop(ctx, client, request)
	storedResponseIDs = append(storedResponseIDs, responseIDs...)
	if err != nil {
		return err
	}

	fmt.Printf("\nStructured result (%s):\n%s\n", response.ID, response.GetOutputText())
	printUsage(response.Usage)

	if err = inspectStoredResponse(ctx, client, response.ID); err != nil {
		return err
	}

	if err = streamFollowUp(ctx, client, config.model, response.ID); err != nil {
		return err
	}

	if config.runBackground {
		background, backgroundErr := runBackgroundResponse(
			ctx,
			client,
			config.model,
			response.ID,
			config.backgroundPrompt,
		)
		if background.ID != "" {
			storedResponseIDs = append(storedResponseIDs, background.ID)
		}
		if backgroundErr != nil {
			return backgroundErr
		}
		fmt.Printf("Background output:\n%s\n", background.GetOutputText())
	}

	if config.runCompaction {
		if err = runStandaloneCompaction(ctx, client, config.model); err != nil {
			return err
		}
	}

	if config.keepResponses {
		fmt.Printf("Stored response IDs retained: %v\n", storedResponseIDs)
	}
	return nil
}

func initialRequest(model string) openai.CreateResponseRequest {
	store := true
	parallelToolCalls := true
	input := []openai.ResponseInputMessage{{
		Role: "user",
		Content: []openai.ResponseInputText{{
			Type: "input_text",
			Text: "Plan one Saturday in London. You must call get_weather for London and use " +
				"web search to find one current museum exhibition. Keep travel practical and cite source URLs.",
		}},
	}}
	instructions := "You are a careful trip planner. Treat get_weather output as simulated example data. " +
		"Use web search for time-sensitive facts, and return the final answer in the requested JSON schema."

	return openai.CreateResponseRequest{
		Model:             model,
		Input:             input,
		Instructions:      instructions,
		MaxOutputTokens:   1200,
		MaxToolCalls:      8,
		Tools:             responseTools(),
		ToolChoice:        "auto",
		ParallelToolCalls: &parallelToolCalls,
		Text:              responseTextConfig(),
		Reasoning:         &openai.ResponseReasoning{Effort: "low", Summary: "auto"},
		Include:           []openai.ResponseInclude{openai.ResponseIncludeWebSearchCallActionSources},
		Store:             &store,
		Truncation:        openai.ResponseTruncationAuto,
		PromptCacheKey:    "go-openai-responses-example-v1",
		SafetyIdentifier:  "responses-example-user",
		Metadata:          map[string]any{"example": "responses", "workflow": "trip-planner"},
	}
}

func inspectStoredResponse(ctx context.Context, client *openai.Client, responseID string) error {
	retrieved, err := client.RetrieveResponse(ctx, responseID, openai.RetrieveResponseOptions{
		Include: []openai.ResponseInclude{openai.ResponseIncludeWebSearchCallActionSources},
	})
	if err != nil {
		return fmt.Errorf("retrieve response %s: %w", responseID, err)
	}
	fmt.Printf("Retrieved stored response: id=%s status=%s output_items=%d\n",
		retrieved.ID, retrieved.Status, len(retrieved.Output))

	items, err := client.ListResponseInputItems(ctx, responseID, openai.ResponseInputItemsListOptions{
		Limit: 100,
		Order: "asc",
	})
	if err != nil {
		return fmt.Errorf("list input items for %s: %w", responseID, err)
	}
	fmt.Printf("Stored input items: %d (has_more=%t)\n", len(items.Data), items.HasMore)
	return nil
}

func parseFlags() settings {
	defaultModel := os.Getenv("OPENAI_MODEL")
	if defaultModel == "" {
		defaultModel = openai.GPT5
	}

	var config settings
	flag.StringVar(&config.model, "model", defaultModel, "Responses API model (or set OPENAI_MODEL)")
	flag.BoolVar(&config.runBackground, "background", false, "also create and poll a background response")
	flag.BoolVar(&config.runCompaction, "compact", false, "also demonstrate standalone context compaction")
	flag.BoolVar(&config.keepResponses, "keep", false, "keep stored responses instead of deleting them")
	flag.DurationVar(&config.requestTimeout, "timeout", 10*time.Minute, "timeout for the complete example")
	flag.StringVar(&config.backgroundPrompt, "background-prompt",
		"Write a detailed, weather-aware packing checklist for this trip.",
		"prompt used by the optional background response")
	flag.Parse()
	return config
}

func responseTools() []openai.ResponseTool {
	weatherParameters := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"location": {
				Type:        jsonschema.String,
				Description: "City and country, for example London, United Kingdom",
			},
			"unit": {
				Type:        jsonschema.String,
				Description: "Temperature unit",
				Enum:        []string{"celsius", "fahrenheit"},
			},
		},
		Required:             []string{"location", "unit"},
		AdditionalProperties: false,
	}

	return []openai.ResponseTool{
		openai.NewResponseFunctionTool(openai.FunctionDefinition{
			Name:        "get_weather",
			Description: "Get simulated weather for a destination used by this example",
			Parameters:  weatherParameters,
			Strict:      true,
		}),
		{
			Type: openai.ToolTypeWebSearch,
			Parameters: map[string]any{
				"search_context_size": "medium",
			},
		},
	}
}

func responseTextConfig() *openai.ResponseTextConfig {
	itineraryItem := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"time":     {Type: jsonschema.String},
			"activity": {Type: jsonschema.String},
		},
		Required:             []string{"time", "activity"},
		AdditionalProperties: false,
	}
	stringItem := jsonschema.Definition{Type: jsonschema.String}
	schema := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"destination":       {Type: jsonschema.String},
			"weather_summary":   {Type: jsonschema.String},
			"current_highlight": {Type: jsonschema.String},
			"itinerary":         {Type: jsonschema.Array, Items: &itineraryItem},
			"sources":           {Type: jsonschema.Array, Items: &stringItem},
		},
		Required: []string{
			"destination",
			"weather_summary",
			"current_highlight",
			"itinerary",
			"sources",
		},
		AdditionalProperties: false,
	}

	return &openai.ResponseTextConfig{
		Verbosity: "medium",
		Format: &openai.ResponseTextFormat{
			Type:        "json_schema",
			Name:        "day_trip_plan",
			Description: "A sourced one-day travel plan",
			Schema:      schema,
			Strict:      true,
		},
	}
}

func createWithToolLoop(
	ctx context.Context,
	client *openai.Client,
	request openai.CreateResponseRequest,
) (openai.CreateResponseResponse, []string, error) {
	var responseIDs []string
	for round := 1; round <= maxToolRounds; round++ {
		response, err := client.CreateResponse(ctx, request)
		if response.ID != "" {
			responseIDs = append(responseIDs, response.ID)
		}
		if err != nil {
			return response, responseIDs, fmt.Errorf("create response (tool round %d): %w", round, err)
		}
		if err = responseError(response); err != nil {
			return response, responseIDs, err
		}

		outputItems, err := decodeOutputItems(response.Output)
		if err != nil {
			return response, responseIDs, err
		}
		printOutputItems(round, outputItems)

		toolOutputs, err := functionCallOutputs(outputItems)
		if err != nil {
			return response, responseIDs, err
		}

		if len(toolOutputs) == 0 {
			if response.Status != openai.ResponseStatusCompleted {
				return response, responseIDs, fmt.Errorf("response %s ended with status %q", response.ID, response.Status)
			}
			return response, responseIDs, nil
		}

		request.PreviousResponseID = response.ID
		request.Input = toolOutputs
	}

	return openai.CreateResponseResponse{}, responseIDs,
		fmt.Errorf("model requested more than %d rounds of local tools", maxToolRounds)
}

func functionCallOutputs(items []openai.ResponseOutputItem) ([]any, error) {
	outputs := make([]any, 0)
	for _, item := range items {
		if item.Type != "function_call" {
			continue
		}
		output, err := callFunction(item)
		if err != nil {
			return nil, err
		}
		fmt.Printf("  local tool %s(%s) -> %s\n", item.Name, item.Arguments, output)
		outputs = append(outputs, openai.ResponseFunctionCallOutput{
			Type:   "function_call_output",
			CallID: item.CallID,
			Output: output,
		})
	}
	return outputs, nil
}

func decodeOutputItems(rawItems []any) ([]openai.ResponseOutputItem, error) {
	items := make([]openai.ResponseOutputItem, 0, len(rawItems))
	for i, rawItem := range rawItems {
		data, err := json.Marshal(rawItem)
		if err != nil {
			return nil, fmt.Errorf("marshal output item %d: %w", i, err)
		}
		var item openai.ResponseOutputItem
		if err = json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("decode output item %d: %w", i, err)
		}
		items = append(items, item)
	}
	return items, nil
}

func printOutputItems(round int, items []openai.ResponseOutputItem) {
	fmt.Printf("Response round %d output items:", round)
	for _, item := range items {
		fmt.Printf(" %s", item.Type)
	}
	fmt.Println()
}

func callFunction(call openai.ResponseOutputItem) (string, error) {
	if call.Name != "get_weather" {
		return "", fmt.Errorf("model requested unknown function %q", call.Name)
	}

	var arguments weatherArguments
	if err := json.Unmarshal([]byte(call.Arguments), &arguments); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", call.Name, err)
	}
	result := weatherResult{
		Location:    arguments.Location,
		Temperature: 17,
		Unit:        arguments.Unit,
		Conditions:  "partly cloudy with a light breeze",
		Source:      "simulated locally by examples/responses",
	}
	if arguments.Unit == "fahrenheit" {
		result.Temperature = 63
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", call.Name, err)
	}
	return string(data), nil
}

func streamFollowUp(ctx context.Context, client *openai.Client, model, previousResponseID string) error {
	store := false
	includeObfuscation := false
	stream, err := client.CreateResponseStream(ctx, openai.CreateResponseRequest{
		Model:              model,
		PreviousResponseID: previousResponseID,
		Input: []openai.ResponseInputMessage{{
			Role: "user",
			Content: []openai.ResponseInputText{{
				Type: "input_text",
				Text: "Summarize the plan in exactly two plain-text sentences for a text message.",
			}},
		}},
		Instructions:  "Be concise. Do not use tools and do not return JSON.",
		Store:         &store,
		Text:          &openai.ResponseTextConfig{Verbosity: "low"},
		StreamOptions: &openai.ResponseStreamOptions{IncludeObfuscation: &includeObfuscation},
	})
	if err != nil {
		return fmt.Errorf("create response stream: %w", err)
	}
	defer stream.Close()

	fmt.Println("\nStreaming follow-up:")
	for {
		event, receiveErr := stream.Recv()
		if errors.Is(receiveErr, io.EOF) {
			fmt.Println()
			return nil
		}
		if receiveErr != nil {
			return fmt.Errorf("receive response stream: %w", receiveErr)
		}

		if event.Type == openai.ResponseStreamEventOutputTextDelta {
			fmt.Print(event.Delta)
			continue
		}
		if event.Type == openai.ResponseStreamEventFailed ||
			event.Type == openai.ResponseStreamEventIncomplete {
			if event.Response != nil {
				return responseError(*event.Response)
			}
			return fmt.Errorf("stream ended with event %q", event.Type)
		}
		if event.Type == openai.ResponseStreamEventError {
			if event.Error != nil {
				return fmt.Errorf("stream error %s: %s", event.Error.Code, event.Error.Message)
			}
			return fmt.Errorf("stream error: %s", event.Message)
		}
	}
}

func runBackgroundResponse(
	ctx context.Context,
	client *openai.Client,
	model string,
	previousResponseID string,
	prompt string,
) (openai.CreateResponseResponse, error) {
	store := true
	response, err := client.CreateResponse(ctx, openai.CreateResponseRequest{
		Model:              model,
		PreviousResponseID: previousResponseID,
		Input:              prompt,
		Instructions:       "Return a thorough but practical checklist in Markdown.",
		Background:         true,
		Store:              &store,
		MaxOutputTokens:    1600,
	})
	if err != nil {
		return response, fmt.Errorf("create background response: %w", err)
	}
	fmt.Printf("\nBackground response %s started with status %s\n", response.ID, response.Status)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for response.Status == openai.ResponseStatusQueued || response.Status == openai.ResponseStatusInProgress {
		select {
		case <-ctx.Done():
			return cancelBackgroundResponse(ctx, client, response, ctx.Err())
		case <-ticker.C:
			response, err = client.RetrieveResponse(ctx, response.ID)
			if err != nil {
				return response, fmt.Errorf("poll background response: %w", err)
			}
			fmt.Printf("Background status: %s\n", response.Status)
		}
	}
	if err = responseError(response); err != nil {
		return response, err
	}
	if response.Status != openai.ResponseStatusCompleted {
		return response, fmt.Errorf("background response %s ended with status %q", response.ID, response.Status)
	}
	return response, nil
}

// cancelBackgroundResponse deliberately starts a short cleanup context because
// the request context has already expired when this function is called.
//
//nolint:contextcheck // Cleanup must outlive the already-cancelled request context.
func cancelBackgroundResponse(
	_ context.Context,
	client *openai.Client,
	response openai.CreateResponseResponse,
	requestErr error,
) (openai.CreateResponseResponse, error) {
	cancelCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cancelled, err := client.CancelResponse(cancelCtx, response.ID)
	if err != nil {
		return response, fmt.Errorf("cancel background response after request ended: %w", err)
	}
	return cancelled, requestErr
}

func runStandaloneCompaction(ctx context.Context, client *openai.Client, model string) error {
	fullContext := []any{
		openai.ResponseInputMessage{
			Role:    "user",
			Content: "Remember that the traveler prefers museums, walking, and vegetarian food.",
		},
		openai.ResponseInputMessage{
			Role:    "assistant",
			Content: "Understood. I will keep those preferences in mind.",
		},
	}
	compacted, err := client.CompactResponse(ctx, openai.CompactResponseRequest{
		Model: model,
		Input: fullContext,
	})
	if err != nil {
		return fmt.Errorf("compact response context: %w", err)
	}
	fmt.Printf("\nStandalone compaction returned %d canonical input items\n", len(compacted.Output))

	nextInput := append([]any(nil), compacted.Output...)
	nextInput = append(nextInput, openai.ResponseInputMessage{
		Role:    "user",
		Content: "In one sentence, restate my travel preferences.",
	})
	store := false
	next, err := client.CreateResponse(ctx, openai.CreateResponseRequest{
		Model: model,
		Input: nextInput,
		Store: &store,
		Text:  &openai.ResponseTextConfig{Verbosity: "low"},
	})
	if err != nil {
		return fmt.Errorf("create response from compacted context: %w", err)
	}
	if err = responseError(next); err != nil {
		return err
	}
	fmt.Printf("Response from compacted context:\n%s\n", next.GetOutputText())
	return nil
}

func responseError(response openai.CreateResponseResponse) error {
	if response.Error != nil {
		return fmt.Errorf("response %s failed (%s): %s", response.ID, response.Error.Code, response.Error.Message)
	}
	if response.IncompleteDetails != nil {
		return fmt.Errorf("response %s is incomplete: %s", response.ID, response.IncompleteDetails.Reason)
	}
	if response.Status == openai.ResponseStatusFailed || response.Status == openai.ResponseStatusCancelled {
		return fmt.Errorf("response %s ended with status %q", response.ID, response.Status)
	}
	return nil
}

func printUsage(usage *openai.ResponseUsage) {
	if usage == nil {
		return
	}
	fmt.Printf("Usage: input=%d output=%d total=%d", usage.InputTokens, usage.OutputTokens, usage.TotalTokens)
	if usage.OutputTokensDetails != nil {
		fmt.Printf(" reasoning=%d", usage.OutputTokensDetails.ReasoningTokens)
	}
	if usage.InputTokensDetails != nil {
		fmt.Printf(" cached=%d", usage.InputTokensDetails.CachedTokens)
	}
	fmt.Println()
}

func deleteStoredResponses(client *openai.Client, responseIDs []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for i := len(responseIDs) - 1; i >= 0; i-- {
		deleted, err := client.DeleteResponse(ctx, responseIDs[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, "delete stored response %s: %v\n", responseIDs[i], err)
			continue
		}
		fmt.Printf("Deleted stored response: %s (deleted=%t)\n", deleted.ID, deleted.Deleted)
	}
}
