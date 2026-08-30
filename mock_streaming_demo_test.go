package openai_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/sashabaranov/go-openai"
)

// This file demonstrates how to create mock clients for go-openai streaming
// functionality. This pattern is useful when testing code that depends on
// go-openai streaming but you want to control the responses for testing.

// MockOpenAIStreamClient demonstrates how to create a full mock client for go-openai.
type MockOpenAIStreamClient struct {
	// Configure canned responses
	ChatCompletionResponse  openai.ChatCompletionResponse
	ChatCompletionStreamErr error

	// Allow function overrides for more complex scenarios
	CreateChatCompletionStreamFn func(
		ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error)
}

func (m *MockOpenAIStreamClient) CreateChatCompletionStream(
	ctx context.Context,
	req openai.ChatCompletionRequest,
) (*openai.ChatCompletionStream, error) {
	if m.CreateChatCompletionStreamFn != nil {
		return m.CreateChatCompletionStreamFn(ctx, req)
	}
	return nil, m.ChatCompletionStreamErr
}

// mockStreamReader creates specific responses for testing.
type mockStreamReader struct {
	responses []openai.ChatCompletionStreamResponse
	index     int
}

func (m *mockStreamReader) Recv() (openai.ChatCompletionStreamResponse, error) {
	if m.index >= len(m.responses) {
		return openai.ChatCompletionStreamResponse{}, io.EOF
	}
	resp := m.responses[m.index]
	m.index++
	return resp, nil
}

func (m *mockStreamReader) Close() error {
	return nil
}

func TestMockOpenAIStreamClient_Demo(t *testing.T) {
	// Create expected responses that our mock stream will return
	expectedResponses := []openai.ChatCompletionStreamResponse{
		{
			ID:     "test-1",
			Object: "chat.completion.chunk",
			Model:  "gpt-3.5-turbo",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Role:    "assistant",
						Content: "Hello",
					},
				},
			},
		},
		{
			ID:     "test-2",
			Object: "chat.completion.chunk",
			Model:  "gpt-3.5-turbo",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: " World",
					},
				},
			},
		},
		{
			ID:     "test-3",
			Object: "chat.completion.chunk",
			Model:  "gpt-3.5-turbo",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index:        0,
					Delta:        openai.ChatCompletionStreamChoiceDelta{},
					FinishReason: "stop",
				},
			},
		},
	}

	// Create mock client with custom stream function
	mockClient := &MockOpenAIStreamClient{
		CreateChatCompletionStreamFn: func(
			_ context.Context, _ openai.ChatCompletionRequest,
		) (*openai.ChatCompletionStream, error) {
			// Create a mock stream reader with our expected responses
			mockStreamReader := &mockStreamReader{
				responses: expectedResponses,
				index:     0,
			}
			// Return a new ChatCompletionStream with our mock reader
			return openai.NewChatCompletionStream(mockStreamReader), nil
		},
	}

	// Test the mock client
	stream, err := mockClient.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("CreateChatCompletionStream returned error: %v", err)
	}
	defer stream.Close()

	// Verify we get back exactly the responses we configured
	fullResponse := ""
	for i, expectedResponse := range expectedResponses {
		receivedResponse, streamErr := stream.Recv()
		if streamErr != nil {
			t.Fatalf("stream.Recv() failed at index %d: %v", i, streamErr)
		}

		// Additional specific checks
		if receivedResponse.ID != expectedResponse.ID {
			t.Errorf("Response %d ID mismatch. Expected: %s, Got: %s",
				i, expectedResponse.ID, receivedResponse.ID)
		}
		if len(receivedResponse.Choices) > 0 && len(expectedResponse.Choices) > 0 {
			expectedContent := expectedResponse.Choices[0].Delta.Content
			receivedContent := receivedResponse.Choices[0].Delta.Content
			if receivedContent != expectedContent {
				t.Errorf("Response %d content mismatch. Expected: %s, Got: %s",
					i, expectedContent, receivedContent)
			}
			fullResponse += receivedContent
		}
	}

	// Verify EOF at the end
	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("Expected EOF at end of stream, got: %v", streamErr)
	}

	// Verify the full assembled response
	expectedFullResponse := "Hello World"
	if fullResponse != expectedFullResponse {
		t.Errorf("Full response mismatch. Expected: %s, Got: %s", expectedFullResponse, fullResponse)
	}

	t.Log("✅ Successfully demonstrated mock OpenAI client with streaming responses!")
	t.Logf("   Full response assembled: %q", fullResponse)
}

// TestMockOpenAIStreamClient_ErrorHandling demonstrates error handling.
func TestMockOpenAIStreamClient_ErrorHandling(t *testing.T) {
	expectedError := errors.New("mock stream error")

	mockClient := &MockOpenAIStreamClient{
		ChatCompletionStreamErr: expectedError,
	}

	_, err := mockClient.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	if !errors.Is(err, expectedError) {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}

	t.Log("✅ Successfully demonstrated mock OpenAI client error handling!")
}
