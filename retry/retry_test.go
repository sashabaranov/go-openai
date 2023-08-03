package retry_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/retry"
)

var ctx = context.Background()

var (
	errCustomError          = errors.New("some custom error")
	errRetryableCustomError = errors.New("my custom retryable error")
)

func TestRetryableLoggerFunc(t *testing.T) {
	t.Parallel()

	result := ""

	policy := retry.Backoff{
		Steps:    3,
		Duration: 0,
		Logger: func(level, msg string) {
			result += fmt.Sprintf("[%s]%s", level, msg)
		},
	}

	_ = retry.OnError(ctx, policy, func() error {
		return context.DeadlineExceeded
	})

	validLogs := "[info]retrying request after 1s to OpenAI 1/3 [context deadline exceeded][info]retrying request after 1s to OpenAI 2/3 [context deadline exceeded][info]retrying request after 1s to OpenAI 3/3 [context deadline exceeded]" //nolint:lll

	if result != validLogs {
		t.Errorf("got [%v], want [%v]", result, validLogs)
	}
}

func TestRetryableErrorFunc(t *testing.T) {
	t.Parallel()

	policy := retry.Backoff{
		RetryableError: func(err error) bool {
			return errors.Is(err, errRetryableCustomError)
		},
	}

	cases := make(map[error]bool)

	cases[errCustomError] = false
	cases[&openai.APIError{HTTPStatusCode: 401}] = false
	cases[errRetryableCustomError] = true

	for k, v := range cases {
		if got := policy.RetryableError(k); v != got {
			t.Errorf("got [%v], want [%v]", v, got)
		}
	}
}

func TestRetry(t *testing.T) {
	t.Parallel()

	canceledContext, cancel := context.WithCancel(ctx)
	cancel()

	testBackoffPolicy := retry.Backoff{
		Steps: 1,
		RetryableError: func(err error) bool {
			return errors.Is(err, errRetryableCustomError)
		},
	}

	type TestCases struct {
		WithCanceledContext bool
		Func                func() error
		ExpectedError       string
	}

	testCases := []TestCases{
		{
			Func: func() error {
				return nil
			},
		},
		{
			ExpectedError: "some custom error",
			Func: func() error {
				return errCustomError
			},
		},
		{
			WithCanceledContext: true,
			ExpectedError:       "context canceled\ncontext error",
			Func: func() error {
				return nil
			},
		},
		{
			ExpectedError: "context deadline exceeded\nretry limit exceeded\nerror making retryable call to OpenAI API",
			Func: func() error {
				return context.DeadlineExceeded
			},
		},
		{
			ExpectedError: "my custom retryable error\nretry limit exceeded\nerror making retryable call to OpenAI API",
			Func: func() error {
				return errRetryableCustomError
			},
		},
		{
			ExpectedError: "error, status code: 401, message: bad auth\ninvalid auth or key",
			Func: func() error {
				return &openai.APIError{
					HTTPStatusCode: http.StatusUnauthorized,
					Message:        "bad auth",
				}
			},
		},
		{
			ExpectedError: "error, status code: 429, message: some too many requests error\nretry limit exceeded\nerror making retryable call to OpenAI API", //nolint:lll
			Func: func() error {
				return &openai.APIError{
					HTTPStatusCode: http.StatusTooManyRequests,
					Message:        "some too many requests error",
				}
			},
		},
		{
			ExpectedError: "error, status code: 500, message: some server error\nretry limit exceeded\nerror making retryable call to OpenAI API", //nolint:lll
			Func: func() error {
				return &openai.APIError{
					HTTPStatusCode: http.StatusInternalServerError,
					Message:        "some server error",
				}
			},
		},
		{
			ExpectedError: "error, status code: 834, message: some status code error\nerror making retryable call to OpenAI API", //nolint:lll
			Func: func() error {
				return &openai.APIError{
					HTTPStatusCode: 834,
					Message:        "some status code error",
				}
			},
		},
	}

	for _, testCase := range testCases {
		testContext := ctx

		if testCase.WithCanceledContext {
			testContext = canceledContext
		}

		err := retry.OnError(testContext, testBackoffPolicy, testCase.Func)

		if len(testCase.ExpectedError) > 0 {
			if err == nil {
				log.Fatal("expected error")
			}

			// test message text
			if err.Error() != testCase.ExpectedError {
				t.Fatalf("got [%s], want [%s]", err.Error(), testCase.ExpectedError)
			}
		}
	}
}
