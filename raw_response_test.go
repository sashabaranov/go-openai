package openai_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

type rawResponseDoer func(*http.Request) (*http.Response, error)

func (do rawResponseDoer) Do(req *http.Request) (*http.Response, error) {
	return do(req)
}

type rawResponseBody struct {
	io.Reader
	closes   int
	closeErr error
}

func (body *rawResponseBody) Close() error {
	body.closes++
	return body.closeErr
}

type rawResponseErrorReader struct {
	err error
}

func (reader rawResponseErrorReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func rawResponseEndpoints() map[string]func(*openai.Client) (openai.RawResponse, error) {
	return map[string]func(*openai.Client) (openai.RawResponse, error){
		"speech": func(client *openai.Client) (openai.RawResponse, error) {
			return client.CreateSpeech(context.Background(), openai.CreateSpeechRequest{
				Model: openai.TTSModel1,
				Input: "Hello",
				Voice: openai.VoiceAlloy,
			})
		},
		"file": func(client *openai.Client) (openai.RawResponse, error) {
			return client.GetFileContent(context.Background(), "file-test")
		},
	}
}

type rawResponseErrorCase struct {
	name    string
	status  int
	content string
	readErr error
}

func TestRawResponseClosesHTTPErrorBody(t *testing.T) {
	readErr := errors.New("response read failed")
	for name, call := range rawResponseEndpoints() {
		t.Run(name, func(t *testing.T) {
			for _, tc := range []rawResponseErrorCase{
				{"api error", http.StatusTooManyRequests,
					`{"error":{"message":"rate limited","type":"rate_limit_error","code":"rate_limit_exceeded"}}`, nil},
				{"non-json error", http.StatusBadGateway, "upstream unavailable", nil},
				{"body read error", http.StatusInternalServerError, "", readErr},
			} {
				t.Run(tc.name, func(t *testing.T) {
					checkRawHTTPErrorBody(t, call, tc)
				})
			}
		})
	}
}

func checkRawHTTPErrorBody(
	t *testing.T,
	call func(*openai.Client) (openai.RawResponse, error),
	tc rawResponseErrorCase,
) {
	t.Helper()
	closeErr := errors.New("response close failed")
	body := &rawResponseBody{Reader: strings.NewReader(tc.content), closeErr: closeErr}
	if tc.readErr != nil {
		body.Reader = rawResponseErrorReader{err: tc.readErr}
	}
	config := openai.DefaultConfig("")
	calls := 0
	config.HTTPClient = rawResponseDoer(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{
			StatusCode: tc.status, Status: http.StatusText(tc.status), Body: body,
		}, nil
	})

	response, err := call(openai.NewClientWithConfig(config))
	if calls != 1 || body.closes != 1 {
		t.Errorf("requests = %d, body closes = %d; want one each", calls, body.closes)
	}
	if response.ReadCloser != nil {
		t.Error("HTTP error response unexpectedly transfers body ownership")
	}
	checkRawResponseError(t, err, tc.status, tc.content, tc.readErr)
	if errors.Is(err, closeErr) {
		t.Error("close error replaced the response error")
	}
}

func checkRawResponseError(t *testing.T, err error, status int, content string, readErr error) {
	t.Helper()
	if readErr != nil {
		if !errors.Is(err, readErr) {
			t.Errorf("error = %v, want wrapped read error", err)
		}
		return
	}
	if status == http.StatusTooManyRequests {
		var apiErr *openai.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("error = %v, want APIError", err)
		}
		if apiErr.HTTPStatusCode != status || apiErr.Message != "rate limited" ||
			apiErr.Type != "rate_limit_error" || apiErr.Code != "rate_limit_exceeded" {
			t.Errorf("unexpected API error: %+v", apiErr)
		}
		return
	}
	var requestErr *openai.RequestError
	if !errors.As(err, &requestErr) {
		t.Fatalf("error = %v, want RequestError", err)
	}
	if requestErr.HTTPStatusCode != status || string(requestErr.Body) != content {
		t.Errorf("unexpected request error: %+v", requestErr)
	}
}

func TestRawResponseSuccessTransfersBody(t *testing.T) {
	for name, call := range rawResponseEndpoints() {
		t.Run(name, func(t *testing.T) {
			body := &rawResponseBody{Reader: strings.NewReader("raw response")}
			config := openai.DefaultConfig("")
			config.HTTPClient = rawResponseDoer(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       body,
					Header:     http.Header{"X-Request-Id": {"request-test"}},
				}, nil
			})
			response, err := call(openai.NewClientWithConfig(config))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if response.ReadCloser != body || body.closes != 0 {
				t.Fatal("success response did not retain the open original body")
			}
			content, err := io.ReadAll(response)
			if err != nil || string(content) != "raw response" {
				t.Errorf("body = %q, error = %v", content, err)
			}
			if response.Header().Get("X-Request-Id") != "request-test" {
				t.Error("success response lost its headers")
			}
			if err = response.Close(); err != nil || body.closes != 1 {
				t.Errorf("caller close error = %v, body closes = %d", err, body.closes)
			}
		})
	}
}

func TestRawResponseTransportError(t *testing.T) {
	transportErr := errors.New("transport failed")
	for name, call := range rawResponseEndpoints() {
		t.Run(name, func(t *testing.T) {
			config := openai.DefaultConfig("")
			config.HTTPClient = rawResponseDoer(func(*http.Request) (*http.Response, error) {
				return nil, transportErr
			})
			response, err := call(openai.NewClientWithConfig(config))
			if !errors.Is(err, transportErr) || response.ReadCloser != nil {
				t.Errorf("response = %+v, error = %v", response, err)
			}
		})
	}
}
