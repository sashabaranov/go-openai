package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/sashabaranov/go-openai/internal/test"
)

var errTestRequestBuilderFailed = errors.New("test request builder failed")

type failingRequestBuilder struct{}

func (*failingRequestBuilder) Build(_ context.Context, _, _ string, _ any) (*http.Request, error) {
	return nil, errTestRequestBuilderFailed
}

func TestClient(t *testing.T) {
	const mockToken = "mock token"
	client := NewClient(mockToken)
	if client.config.authToken != mockToken {
		t.Errorf("Client does not contain proper token")
	}

	const mockOrg = "mock org"
	client = NewOrgClient(mockToken, mockOrg)
	if client.config.authToken != mockToken {
		t.Errorf("Client does not contain proper token")
	}
	if client.config.OrgID != mockOrg {
		t.Errorf("Client does not contain proper orgID")
	}
}

func TestDecodeResponse(t *testing.T) {
	stringInput := ""

	testCases := []struct {
		name  string
		value interface{}
		body  io.Reader
	}{
		{
			name:  "nil input",
			value: nil,
			body:  bytes.NewReader([]byte("")),
		},
		{
			name:  "string input",
			value: &stringInput,
			body:  bytes.NewReader([]byte("test")),
		},
		{
			name:  "map input",
			value: &map[string]interface{}{},
			body:  bytes.NewReader([]byte(`{"test": "test"}`)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := decodeResponse(tc.body, tc.value)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestHandleErrorResp(t *testing.T) {
	// var errRes *ErrorResponse
	var errRes ErrorResponse
	var reqErr RequestError
	t.Log(errRes, errRes.Error)
	if errRes.Error != nil {
		reqErr.Err = errRes.Error
	}
	t.Log(fmt.Errorf("error, %w", &reqErr))
	t.Log(errRes.Error, "nil pointer check Pass")

	const mockToken = "mock token"
	client := NewClient(mockToken)

	testCases := []struct {
		name     string
		httpCode int
		body     io.Reader
		expected string
	}{
		{
			name:     "401 Invalid Authentication",
			httpCode: http.StatusUnauthorized,
			body: bytes.NewReader([]byte(
				`{
					"error":{
						"message":"You didn't provide an API key. ....",
						"type":"invalid_request_error",
						"param":null,
						"code":null
					}
				}`,
			)),
			expected: "error, status code: 401, message: You didn't provide an API key. ....",
		},
		{
			name:     "401 Azure Access Denied",
			httpCode: http.StatusUnauthorized,
			body: bytes.NewReader([]byte(
				`{
					"error":{
						"code":"AccessDenied",
						"message":"Access denied due to Virtual Network/Firewall rules."
					}
				}`,
			)),
			expected: "error, status code: 401, message: Access denied due to Virtual Network/Firewall rules.",
		},
		{
			name:     "503 Model Overloaded",
			httpCode: http.StatusServiceUnavailable,
			body: bytes.NewReader([]byte(`
				{
					"error":{
						"message":"That model...",
						"type":"server_error",
						"param":null,
						"code":null
					}
				}`)),
			expected: "error, status code: 503, message: That model...",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCase := &http.Response{}
			testCase.StatusCode = tc.httpCode
			testCase.Body = io.NopCloser(tc.body)
			err := client.handleErrorResp(testCase)
			t.Log(err.Error())
			if err.Error() != tc.expected {
				t.Errorf("Unexpected error: %v , expected: %s", err, tc.expected)
				t.Fail()
			}

			e := &APIError{}
			if !errors.As(err, &e) {
				t.Errorf("(%s) Expected error to be of type APIError", tc.name)
				t.Fail()
			}
		})
	}
}

func TestClientReturnsRequestBuilderErrors(t *testing.T) {
	var err error
	ts := test.NewTestServer().OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	client.requestBuilder = &failingRequestBuilder{}

	ctx := context.Background()

	_, err = client.CreateCompletion(ctx, CompletionRequest{Prompt: "testing"})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateChatCompletion(ctx, ChatCompletionRequest{Model: GPT3Dot5Turbo})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateChatCompletionStream(ctx, ChatCompletionRequest{Model: GPT3Dot5Turbo})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateFineTune(ctx, FineTuneRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListFineTunes(ctx)
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CancelFineTune(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.GetFineTune(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.DeleteFineTune(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListFineTuneEvents(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.Moderations(ctx, ModerationRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.Edits(ctx, EditsRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateEmbeddings(ctx, EmbeddingRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateImage(ctx, ImageRequest{})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	err = client.DeleteFile(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.GetFile(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListFiles(ctx)
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListEngines(ctx)
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.GetEngine(ctx, "")
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.ListModels(ctx)
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateCompletionStream(ctx, CompletionRequest{Prompt: ""})
	if !errors.Is(err, errTestRequestBuilderFailed) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}
}

func TestClientReturnsRequestBuilderErrorsAddtion(t *testing.T) {
	var err error
	ts := test.NewTestServer().OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	client.requestBuilder = &failingRequestBuilder{}

	ctx := context.Background()

	_, err = client.CreateCompletion(ctx, CompletionRequest{Prompt: 1})
	if !errors.Is(err, ErrCompletionRequestPromptTypeNotSupported) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}

	_, err = client.CreateCompletionStream(ctx, CompletionRequest{Prompt: 1})
	if !errors.Is(err, ErrCompletionRequestPromptTypeNotSupported) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}
}
