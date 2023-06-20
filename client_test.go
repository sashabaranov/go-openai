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

func (*failingRequestBuilder) Build(_ context.Context, _, _ string, _ any, _ http.Header) (*http.Request, error) {
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
		name     string
		value    interface{}
		body     io.Reader
		hasError bool
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
		{
			name:     "reader return error",
			value:    &stringInput,
			body:     &errorReader{err: errors.New("dummy")},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := decodeResponse(tc.body, tc.value)
			if (err != nil) != tc.hasError {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

type errorReader struct {
	err error
}

func (e *errorReader) Read(_ []byte) (n int, err error) {
	return 0, e.err
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
		{
			name:     "503 no message (Unknown response)",
			httpCode: http.StatusServiceUnavailable,
			body: bytes.NewReader([]byte(`
				{
					"error":{}
				}`)),
			expected: "error, status code: 503, message: ",
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
	config := DefaultConfig(test.GetTestToken())
	client := NewClientWithConfig(config)
	client.requestBuilder = &failingRequestBuilder{}
	ctx := context.Background()

	type TestCase struct {
		Name     string
		TestFunc func() (any, error)
	}

	testCases := []TestCase{
		{"CreateCompletion", func() (any, error) {
			return client.CreateCompletion(ctx, CompletionRequest{Prompt: "testing"})
		}},
		{"CreateCompletionStream", func() (any, error) {
			return client.CreateCompletionStream(ctx, CompletionRequest{Prompt: ""})
		}},
		{"CreateChatCompletion", func() (any, error) {
			return client.CreateChatCompletion(ctx, ChatCompletionRequest{Model: GPT3Dot5Turbo})
		}},
		{"CreateChatCompletionStream", func() (any, error) {
			return client.CreateChatCompletionStream(ctx, ChatCompletionRequest{Model: GPT3Dot5Turbo})
		}},
		{"CreateFineTune", func() (any, error) {
			return client.CreateFineTune(ctx, FineTuneRequest{})
		}},
		{"ListFineTunes", func() (any, error) {
			return client.ListFineTunes(ctx)
		}},
		{"CancelFineTune", func() (any, error) {
			return client.CancelFineTune(ctx, "")
		}},
		{"GetFineTune", func() (any, error) {
			return client.GetFineTune(ctx, "")
		}},
		{"DeleteFineTune", func() (any, error) {
			return client.DeleteFineTune(ctx, "")
		}},
		{"ListFineTuneEvents", func() (any, error) {
			return client.ListFineTuneEvents(ctx, "")
		}},
		{"Moderations", func() (any, error) {
			return client.Moderations(ctx, ModerationRequest{})
		}},
		{"Edits", func() (any, error) {
			return client.Edits(ctx, EditsRequest{})
		}},
		{"CreateEmbeddings", func() (any, error) {
			return client.CreateEmbeddings(ctx, EmbeddingRequest{})
		}},
		{"CreateImage", func() (any, error) {
			return client.CreateImage(ctx, ImageRequest{})
		}},
		{"DeleteFile", func() (any, error) {
			return nil, client.DeleteFile(ctx, "")
		}},
		{"GetFile", func() (any, error) {
			return client.GetFile(ctx, "")
		}},
		{"GetFileContent", func() (any, error) {
			return client.GetFileContent(ctx, "")
		}},
		{"ListFiles", func() (any, error) {
			return client.ListFiles(ctx)
		}},
		{"ListEngines", func() (any, error) {
			return client.ListEngines(ctx)
		}},
		{"GetEngine", func() (any, error) {
			return client.GetEngine(ctx, "")
		}},
		{"ListModels", func() (any, error) {
			return client.ListModels(ctx)
		}},
		{"GetModel", func() (any, error) {
			return client.GetModel(ctx, "text-davinci-003")
		}},
	}

	for _, testCase := range testCases {
		_, err := testCase.TestFunc()
		if !errors.Is(err, errTestRequestBuilderFailed) {
			t.Fatalf("%s did not return error when request builder failed: %v", testCase.Name, err)
		}
	}
}

func TestClientReturnsRequestBuilderErrorsAddtion(t *testing.T) {
	config := DefaultConfig(test.GetTestToken())
	client := NewClientWithConfig(config)
	client.requestBuilder = &failingRequestBuilder{}
	ctx := context.Background()
	_, err := client.CreateCompletion(ctx, CompletionRequest{Prompt: 1})
	if !errors.Is(err, ErrCompletionRequestPromptTypeNotSupported) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}
	_, err = client.CreateCompletionStream(ctx, CompletionRequest{Prompt: 1})
	if !errors.Is(err, ErrCompletionRequestPromptTypeNotSupported) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}
}
