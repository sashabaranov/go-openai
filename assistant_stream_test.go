package openai_test

import (
	"context"
	"io"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestAssistantStream(t *testing.T) {
	theadID := "thread_123"
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/threads",
		func(w http.ResponseWriter, _ *http.Request) {
			resBytes, _ := json.Marshal(openai.Thread{
				ID:        theadID,
				Object:    "thread",
				CreatedAt: 1699012949,
			})
			fmt.Fprintln(w, string(resBytes))
		})

	server.RegisterHandler("/v1/threads/"+theadID+"/runs",
		func(w http.ResponseWriter, _ *http.Request) {
			//nolint:lll
			resBytes := []byte(`
event: thread.run.created
data: {"id":"run_123","object":"thread.run","created_at":1710330640,"assistant_id":"asst_123","thread_id":"thread_123","status":"queued","started_at":null,"expires_at":1710331240,"cancelled_at":null,"failed_at":null,"completed_at":null,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[],"file_ids":[],"metadata":{},"usage":null}

event: thread.run.queued
data: {"id":"run_123","object":"thread.run","created_at":1710330640,"assistant_id":"asst_123","thread_id":"thread_123","status":"queued","started_at":null,"expires_at":1710331240,"cancelled_at":null,"failed_at":null,"completed_at":null,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[],"file_ids":[],"metadata":{},"usage":null}

event: thread.run.in_progress
data: {"id":"run_123","object":"thread.run","created_at":1710330640,"assistant_id":"asst_123","thread_id":"thread_123","status":"in_progress","started_at":1710330641,"expires_at":1710331240,"cancelled_at":null,"failed_at":null,"completed_at":null,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[],"file_ids":[],"metadata":{},"usage":null}

event: thread.run.step.created
data: {"id":"step_001","object":"thread.run.step","created_at":1710330641,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"message_creation","status":"in_progress","cancelled_at":null,"completed_at":null,"expires_at":1710331240,"failed_at":null,"last_error":null,"step_details":{"type":"message_creation","message_creation":{"message_id":"msg_001"}},"usage":null}

event: thread.run.step.in_progress
data: {"id":"step_001","object":"thread.run.step","created_at":1710330641,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"message_creation","status":"in_progress","cancelled_at":null,"completed_at":null,"expires_at":1710331240,"failed_at":null,"last_error":null,"step_details":{"type":"message_creation","message_creation":{"message_id":"msg_001"}},"usage":null}

event: thread.message.created
data: {"id":"msg_001","object":"thread.message","created_at":1710330641,"assistant_id":"asst_123","thread_id":"thread_123","run_id":"run_123","status":"in_progress","incomplete_details":null,"incomplete_at":null,"completed_at":null,"role":"assistant","content":[],"file_ids":[],"metadata":{}}

event: thread.message.in_progress
data: {"id":"msg_001","object":"thread.message","created_at":1710330641,"assistant_id":"asst_123","thread_id":"thread_123","run_id":"run_123","status":"in_progress","incomplete_details":null,"incomplete_at":null,"completed_at":null,"role":"assistant","content":[],"file_ids":[],"metadata":{}}

event: thread.message.delta
data: {"id":"msg_001","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"Hello","annotations":[]}}]}}

event: thread.message.delta
data: {"id":"msg_001","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":" today"}}]}}

event: thread.message.delta
data: {"id":"msg_001","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"?"}}]}}

event: thread.message.completed
data: {"id":"msg_001","object":"thread.message","created_at":1710330641,"assistant_id":"asst_123","thread_id":"thread_123","run_id":"run_123","status":"completed","incomplete_details":null,"incomplete_at":null,"completed_at":1710330642,"role":"assistant","content":[{"type":"text","text":{"value":"Hello! How can I assist you today?","annotations":[]}}],"file_ids":[],"metadata":{}}

event: thread.run.step.completed
data: {"id":"step_001","object":"thread.run.step","created_at":1710330641,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"message_creation","status":"completed","cancelled_at":null,"completed_at":1710330642,"expires_at":1710331240,"failed_at":null,"last_error":null,"step_details":{"type":"message_creation","message_creation":{"message_id":"msg_001"}},"usage":{"prompt_tokens":20,"completion_tokens":11,"total_tokens":31}}

event: thread.run.completed
data: {"id":"run_123","object":"thread.run","created_at":1710330640,"assistant_id":"asst_123","thread_id":"thread_123","status":"completed","started_at":1710330641,"expires_at":null,"cancelled_at":null,"failed_at":null,"completed_at":1710330642,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[],"file_ids":[],"metadata":{},"usage":{"prompt_tokens":20,"completion_tokens":11,"total_tokens":31}}

event: done
data: [DONE]
`)
			fmt.Fprintln(w, string(resBytes))
		})

	ctx := context.Background()

	thread, err := client.CreateThread(ctx, openai.ThreadRequest{
		Messages: []openai.ThreadMessage{{Role: openai.ThreadMessageRoleUser, Content: "i want to go home"}},
	})

	if err != nil {
		t.Fatalf("CreateThread failed, error: %v", err)
	}

	stream, err := client.CreateAssistantThreadRunStream(ctx, thread.ID, openai.RunRequest{
		AssistantID: "asst_123",
		Model:       openai.GPT4TurboPreview,
	})

	checks.NoError(t, err, "CreateAssistentStream failed")

	defer stream.Close()

	var messageDelta string
	completed := false

	err = stream.On("thread.message.delta", func(resp openai.AssistantThreadRunStreamResponse, _ []byte) {
		messageDelta += resp.Delta.Content[0].Text.Value
	})

	if err != nil {
		t.Fatalf("Stream set event callback failed, error: %v", err)
	}

	err = stream.On("thread.message.completed", func(_ openai.AssistantThreadRunStreamResponse, _ []byte) {
		completed = true
	})

	if err != nil {
		t.Fatalf("Stream set event callback failed, error: %v", err)
	}

	err = stream.Run()

	checks.ErrorIs(t, err, io.EOF, "Stream run error", err.Error())

	if !completed || messageDelta != "Hello today?" {
		t.Fatalf("Wrong message: %s", messageDelta)
	}
}

func TestCreatAssistantStreamError(t *testing.T) {
	client, _, teardown := setupOpenAITestServer()
	defer teardown()

	_, err := client.CreateAssistantThreadRunStream(context.Background(), "thread_123", openai.RunRequest{
		AssistantID: "asst_123",
		Model:       openai.GPT4TurboPreview,
	})

	checks.HasError(t, err, "CreateAssistentStream should return error")
}

func TestAssistantStreamWithTools(t *testing.T) {
	theadID := "thread_123"
	runID := "run_123"
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/threads/"+theadID+"/runs/"+runID+"/submit_tool_outputs",
		func(w http.ResponseWriter, _ *http.Request) {
			//nolint:lll
			resBytes := []byte(`
event: thread.run.step.completed
data: {"id":"step_001","object":"thread.run.step","created_at":1710352449,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"tool_calls","status":"completed","cancelled_at":null,"completed_at":1710352475,"expires_at":1710353047,"failed_at":null,"last_error":null,"step_details":{"type":"tool_calls","tool_calls":[{"id":"call_iWr0kQ2EaYMaxNdl0v3KYkx7","type":"function","function":{"name":"get_current_weather","arguments":"{\"location\":\"San Francisco, CA\",\"unit\":\"fahrenheit\"}","output":"70 degrees and sunny."}}]},"usage":{"prompt_tokens":291,"completion_tokens":24,"total_tokens":315}}

event: thread.run.queued
data: {"id":"run_123","object":"thread.run","created_at":1710352447,"assistant_id":"asst_123","thread_id":"thread_123","status":"queued","started_at":1710352448,"expires_at":1710353047,"cancelled_at":null,"failed_at":null,"completed_at":null,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[{"type":"function","function":{"name":"get_current_weather","description":"Get the current weather in a given location","parameters":{"type":"object","properties":{"location":{"type":"string","description":"The city and state, e.g. San Francisco, CA"},"unit":{"type":"string","enum":["celsius","fahrenheit"]}},"required":["location"]}}}],"file_ids":[],"metadata":{},"usage":null}

event: thread.run.in_progress
data: {"id":"run_123","object":"thread.run","created_at":1710352447,"assistant_id":"asst_123","thread_id":"thread_123","status":"in_progress","started_at":1710352475,"expires_at":1710353047,"cancelled_at":null,"failed_at":null,"completed_at":null,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[{"type":"function","function":{"name":"get_current_weather","description":"Get the current weather in a given location","parameters":{"type":"object","properties":{"location":{"type":"string","description":"The city and state, e.g. San Francisco, CA"},"unit":{"type":"string","enum":["celsius","fahrenheit"]}},"required":["location"]}}}],"file_ids":[],"metadata":{},"usage":null}

event: thread.run.step.created
data: {"id":"step_002","object":"thread.run.step","created_at":1710352476,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"message_creation","status":"in_progress","cancelled_at":null,"completed_at":null,"expires_at":1710353047,"failed_at":null,"last_error":null,"step_details":{"type":"message_creation","message_creation":{"message_id":"msg_002"}},"usage":null}

event: thread.run.step.in_progress
data: {"id":"step_002","object":"thread.run.step","created_at":1710352476,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"message_creation","status":"in_progress","cancelled_at":null,"completed_at":null,"expires_at":1710353047,"failed_at":null,"last_error":null,"step_details":{"type":"message_creation","message_creation":{"message_id":"msg_002"}},"usage":null}

event: thread.message.created
data: {"id":"msg_002","object":"thread.message","created_at":1710352476,"assistant_id":"asst_123","thread_id":"thread_123","run_id":"run_123","status":"in_progress","incomplete_details":null,"incomplete_at":null,"completed_at":null,"role":"assistant","content":[],"file_ids":[],"metadata":{}}

event: thread.message.in_progress
data: {"id":"msg_002","object":"thread.message","created_at":1710352476,"assistant_id":"asst_123","thread_id":"thread_123","run_id":"run_123","status":"in_progress","incomplete_details":null,"incomplete_at":null,"completed_at":null,"role":"assistant","content":[],"file_ids":[],"metadata":{}}

event: thread.message.delta
data: {"id":"msg_002","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"The","annotations":[]}}]}}

event: thread.message.delta
data: {"id":"msg_002","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":" current"}}]}}

event: thread.message.delta
data: {"id":"msg_002","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":" weather"}}]}}

event: thread.message.delta
data: {"id":"msg_002","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":" sunny"}}]}}

event: thread.message.delta
data: {"id":"msg_002","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"."}}]}}

event: thread.message.completed
data: {"id":"msg_002","object":"thread.message","created_at":1710352476,"assistant_id":"asst_123","thread_id":"thread_123","run_id":"run_123","status":"completed","incomplete_details":null,"incomplete_at":null,"completed_at":1710352477,"role":"assistant","content":[{"type":"text","text":{"value":"The current weather in San Francisco, CA is 70 degrees Fahrenheit and sunny.","annotations":[]}}],"file_ids":[],"metadata":{}}

event: thread.run.step.completed
data: {"id":"step_002","object":"thread.run.step","created_at":1710352476,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"message_creation","status":"completed","cancelled_at":null,"completed_at":1710352477,"expires_at":1710353047,"failed_at":null,"last_error":null,"step_details":{"type":"message_creation","message_creation":{"message_id":"msg_002"}},"usage":{"prompt_tokens":329,"completion_tokens":18,"total_tokens":347}}

event: thread.run.completed
data: {"id":"run_123","object":"thread.run","created_at":1710352447,"assistant_id":"asst_123","thread_id":"thread_123","status":"completed","started_at":1710352475,"expires_at":null,"cancelled_at":null,"failed_at":null,"completed_at":1710352477,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[{"type":"function","function":{"name":"get_current_weather","description":"Get the current weather in a given location","parameters":{"type":"object","properties":{"location":{"type":"string","description":"The city and state, e.g. San Francisco, CA"},"unit":{"type":"string","enum":["celsius","fahrenheit"]}},"required":["location"]}}}],"file_ids":[],"metadata":{},"usage":{"prompt_tokens":620,"completion_tokens":42,"total_tokens":662}}

event: done
data: [DONE]
`)
			fmt.Fprintln(w, string(resBytes))
		})

	server.RegisterHandler("/v1/threads",
		func(w http.ResponseWriter, _ *http.Request) {
			resBytes, _ := json.Marshal(openai.Thread{
				ID:        theadID,
				Object:    "thread",
				CreatedAt: 1699012949,
			})
			fmt.Fprintln(w, string(resBytes))
		})

	server.RegisterHandler("/v1/threads/"+theadID+"/runs",
		func(w http.ResponseWriter, _ *http.Request) {
			//nolint:lll
			resBytes := []byte(`
event: thread.run.created
data: {"id":"run_123","object":"thread.run","created_at":1710330640,"assistant_id":"asst_123","thread_id":"thread_123","status":"queued","started_at":null,"expires_at":1710331240,"cancelled_at":null,"failed_at":null,"completed_at":null,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[],"file_ids":[],"metadata":{},"usage":null}

event: thread.run.queued
data: {"id":"run_123","object":"thread.run","created_at":1710330640,"assistant_id":"asst_123","thread_id":"thread_123","status":"queued","started_at":null,"expires_at":1710331240,"cancelled_at":null,"failed_at":null,"completed_at":null,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[],"file_ids":[],"metadata":{},"usage":null}

event: thread.run.in_progress
data: {"id":"run_123","object":"thread.run","created_at":1710330640,"assistant_id":"asst_123","thread_id":"thread_123","status":"in_progress","started_at":1710330641,"expires_at":1710331240,"cancelled_at":null,"failed_at":null,"completed_at":null,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[],"file_ids":[],"metadata":{},"usage":null}

event: thread.run.step.created
data: {"id":"step_001","object":"thread.run.step","created_at":1710330641,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"message_creation","status":"in_progress","cancelled_at":null,"completed_at":null,"expires_at":1710331240,"failed_at":null,"last_error":null,"step_details":{"type":"message_creation","message_creation":{"message_id":"msg_001"}},"usage":null}

event: thread.run.step.in_progress
data: {"id":"step_001","object":"thread.run.step","created_at":1710330641,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"message_creation","status":"in_progress","cancelled_at":null,"completed_at":null,"expires_at":1710331240,"failed_at":null,"last_error":null,"step_details":{"type":"message_creation","message_creation":{"message_id":"msg_001"}},"usage":null}

event: thread.message.created
data: {"id":"msg_001","object":"thread.message","created_at":1710330641,"assistant_id":"asst_123","thread_id":"thread_123","run_id":"run_123","status":"in_progress","incomplete_details":null,"incomplete_at":null,"completed_at":null,"role":"assistant","content":[],"file_ids":[],"metadata":{}}

event: thread.message.in_progress
data: {"id":"msg_001","object":"thread.message","created_at":1710330641,"assistant_id":"asst_123","thread_id":"thread_123","run_id":"run_123","status":"in_progress","incomplete_details":null,"incomplete_at":null,"completed_at":null,"role":"assistant","content":[],"file_ids":[],"metadata":{}}

event: thread.message.delta
data: {"id":"msg_001","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"Hello","annotations":[]}}]}}

event: thread.message.delta
data: {"id":"msg_001","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":" today"}}]}}

event: thread.message.delta
data: {"id":"msg_001","object":"thread.message.delta","delta":{"content":[{"index":0,"type":"text","text":{"value":"?"}}]}}

event: thread.run.requires_action
data: { "id": "run_123", "object": "thread.run", "assistant_id": "asst_123", "thread_id": "thread_123", "status": "requires_action", "required_action": { "type": "submit_tool_outputs", "submit_tool_outputs": { "tool_calls": [ { "id": "call_123", "type": "function", "function": { "name": "getCurrentWeather", "arguments": "{\"location\":\"San Francisco\"}" } }, { "id": "call_456", "type": "function", "function": { "name": "getNickname", "arguments": "{\"location\":\"Los Angeles\"}" } } ] } } }

event: thread.message.completed
data: {"id":"msg_001","object":"thread.message","created_at":1710330641,"assistant_id":"asst_123","thread_id":"thread_123","run_id":"run_123","status":"completed","incomplete_details":null,"incomplete_at":null,"completed_at":1710330642,"role":"assistant","content":[{"type":"text","text":{"value":"Hello! How can I assist you today?","annotations":[]}}],"file_ids":[],"metadata":{}}

event: thread.run.step.completed
data: {"id":"step_001","object":"thread.run.step","created_at":1710330641,"run_id":"run_123","assistant_id":"asst_123","thread_id":"thread_123","type":"message_creation","status":"completed","cancelled_at":null,"completed_at":1710330642,"expires_at":1710331240,"failed_at":null,"last_error":null,"step_details":{"type":"message_creation","message_creation":{"message_id":"msg_001"}},"usage":{"prompt_tokens":20,"completion_tokens":11,"total_tokens":31}}

event: thread.run.completed
data: {"id":"run_123","object":"thread.run","created_at":1710330640,"assistant_id":"asst_123","thread_id":"thread_123","status":"completed","started_at":1710330641,"expires_at":null,"cancelled_at":null,"failed_at":null,"completed_at":1710330642,"required_action":null,"last_error":null,"model":"gpt-4-turbo","instructions":null,"tools":[],"file_ids":[],"metadata":{},"usage":{"prompt_tokens":20,"completion_tokens":11,"total_tokens":31}}

event: done
data: [DONE]
`)
			fmt.Fprintln(w, string(resBytes))
		})

	ctx := context.Background()

	thread, err := client.CreateThread(ctx, openai.ThreadRequest{
		Messages: []openai.ThreadMessage{{Role: openai.ThreadMessageRoleUser, Content: "How's the weather in San Francisco"}},
	})

	if err != nil {
		t.Fatalf("CreateThread failed, error: %v", err)
	}

	var tool openai.Tool
	err = json.Unmarshal([]byte(`
{
    "type": "function",
    "function": {
        "name": "getCurrentWeather",
        "description": "Get the current weather in a given location",
        "parameters": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "The city and state, e.g. San Francisco, CA"
                },
                "unit": {
                    "type": "string",
                    "enum": [
                        "celsius",
                        "fahrenheit"
                    ]
                }
            },
            "required": [
                "location"
            ]
        }
    }
}
`), &tool)

	if err != nil {
		t.Fatalf("Tool Unmarshal failed, error: %v", err)
	}

	stream, err := client.CreateAssistantThreadRunStream(ctx, thread.ID, openai.RunRequest{
		AssistantID: "asst_123",
		Model:       openai.GPT4TurboPreview,
		Tools:       []openai.Tool{tool},
	})

	checks.NoError(t, err, "CreateAssistentStream failed")

	var messageDelta string
	var requiredAction *openai.RunRequiredAction

	err = stream.On("thread.run.requires_action", func(_ openai.AssistantThreadRunStreamResponse, rawData []byte) {
		runStep := openai.Run{}
		err = json.Unmarshal(rawData, &runStep)
		if err != nil {
			t.Fatalf("Thread Run Unmarshal failed, error: %v", err)
		}
		requiredAction = runStep.RequiredAction
		runID = runStep.ID
	})

	if err != nil {
		t.Fatalf("Stream set event callback failed, error: %v", err)
	}

	err = stream.Run()

	checks.ErrorIs(t, err, io.EOF, "Stream run error", err.Error())

	toolOuputs := []openai.ToolOutput{}
	for _, call := range requiredAction.SubmitToolOutputs.ToolCalls {
		output := openai.ToolOutput{
			ToolCallID: call.ID,
			Output:     "70 degrees and sunny.",
		}
		toolOuputs = append(toolOuputs, output)
	}

	if len(toolOuputs) < 1 {
		t.Fatalf("ToolOutputs error, %v", toolOuputs)
	}

	stream, err = client.CreateAssistantThreadRunToolStream(ctx, theadID, runID, openai.SubmitToolOutputsRequest{
		ToolOutputs: toolOuputs,
	})

	checks.NoError(t, err, "CreateAssistentStream failed")

	err = stream.On("thread.message.delta", func(resp openai.AssistantThreadRunStreamResponse, _ []byte) {
		messageDelta += resp.Delta.Content[0].Text.Value
	})

	if err != nil {
		t.Fatalf("Stream set event callback failed, error: %v", err)
	}

	err = stream.Run()

	checks.ErrorIs(t, err, io.EOF, "Stream run error", err.Error())

	if messageDelta != "The current weather sunny." {
		t.Fatalf("Wrong message: %s", messageDelta)
	}
}

func TestAssistantStreamWithToolsError(t *testing.T) {
	client, _, teardown := setupOpenAITestServer()
	defer teardown()

	//nolint:lll
	_, err := client.CreateAssistantThreadRunToolStream(context.Background(), "thread_123", "run_123", openai.SubmitToolOutputsRequest{})

	checks.HasError(t, err, "CreateAssistantThreadRunToolStream should return error")
}
