package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
			name:     "429 Request Throttled Too Many Requests",
			httpCode: http.StatusTooManyRequests,
			body: bytes.NewReader([]byte(
				`{
					"error":{
						"code":"429",
						"message":"That model..."
					}
				}`,
			)),
			expected: "error, status code: 429, message: That model...",
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

func TestRequestImageErrors(t *testing.T) {
	config := DefaultAzureConfig(test.GetTestToken(), "http://localhost:8080/openai/operations/images")
	client := NewClientWithConfig(config)
	// Test requestImage callback URL empty.
	testCase := "Callback URL is empty"
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
	}
	v := &ImageRequest{}
	err := client.requestImage(res, v)
	if !errors.Is(err, ErrClientEmptyCallbackURL) {
		t.Fatalf("%s did not return error. requestImage failed: %v", testCase, err)
	}
	// Test requestImage callback URL malformed.
	testCase = "Callback URL is malformed"
	res = &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Operation-Location": []string{"hxxp://localhost:8080/openai/operations/images/request-id"}},
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
	}
	err = client.requestImage(res, v)
	if err == nil {
		t.Fatalf("%s did not return error. requestImage failed: %v", testCase, err)
	}
	// Test requestImage callback URL with invalid chars that cause url.Parse to throw error.
	testCase = "Callback URL fails URL parsing"
	res = &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Operation-Location": []string{"http://abc{DEf1=ghi@localhost:8080/openai/operations/images/request-id"}},
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
	}
	err = client.requestImage(res, v)
	if err == nil {
		t.Fatalf("%s did not return error. requestImage failed: %v", testCase, err)
	}
}

func TestImageRequestCallbackErrors(t *testing.T) {
	config := DefaultAzureConfig(test.GetTestToken(), "http://localhost:8080/openai/operations/images")
	client := NewClientWithConfig(config)
	// Test imageRequestCallback status response empty.
	testCase := "imageRequestCallback status response empty"
	var request ImageRequest
	ctx := context.Background()
	req, err := client.requestBuilder.Build(ctx, http.MethodPost, client.fullURL("openai/operations/images"), request)
	if err != nil {
		t.Fatalf("%s. requestBuilder failed with unexpected error: %v", testCase, err)
	}
	cbResponse := CallBackResponse{
		Created: time.Now().Unix(),
		Status:  "",
		Result: CBResult{
			Data: CBData{
				{URL: "http://example.com/image1"},
				{URL: "http://example.com/image2"},
			},
		},
	}
	cbResponseBytes := new(bytes.Buffer)
	err = json.NewEncoder(cbResponseBytes).Encode(cbResponse)
	if err != nil {
		t.Fatalf("%s. json encoding failed with unexpected error: %v", testCase, err)
	}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString(cbResponseBytes.String())),
	}
	v := &ImageRequest{}
	err = client.imageRequestCallback(req, v, res)

	if !errors.Is(err, ErrClientRetrievingCallbackResponse) {
		t.Fatalf("%s did not return error. imageRequestCallback failed: %v", testCase, err)
	}
}

func TestRequestImageFunc(t *testing.T) {
	config := DefaultAzureConfig(test.GetTestToken(), "http://localhost:8080/openai/operations/images")
	client := NewClientWithConfig(config)
	v := &ImageRequest{}
	var errorHTTPClient httptest.ResponseRecorder
	errorHTTPClient.WriteHeader(http.StatusInternalServerError)
	err := client.requestImage(errorHTTPClient.Result(), v)
	if err == nil {
		t.Fatalf("%s. requestBuilder failed with unexpected error: %v", "TestRequestImageFunc", err)
	}
}
