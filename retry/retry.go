package retry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
)

var (
	ErrRetryLimitExceeded  = errors.New("retry limit exceeded")
	ErrInvalidAuth         = errors.New("invalid auth or key")
	ErrMakingRetryableCall = errors.New("error making retryable call to OpenAI API")
	ErrContextError        = errors.New("context error")
	ErrNotAnError          = errors.New("not an error")
)

type Backoff struct {
	Steps          int
	Duration       time.Duration
	RetryableError func(error) bool
	Logger         func(level, msg string)
}

func (b *Backoff) CanRetry(ctx context.Context, currentTry int, err error, d time.Duration) error {
	if currentTry > b.Steps {
		return errors.Join(err, ErrRetryLimitExceeded)
	}

	if b.Logger != nil {
		b.Logger("info", fmt.Sprintf("retrying request after %s to OpenAI %d/%d [%s]", d.String(), currentTry, b.Steps, err.Error())) //nolint:lll
	}

	// wait before retry
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}

	return nil
}

const (
	defaultRetrySteps    = 10
	defaultRetryDuration = 5 * time.Second
)

var DefaultRetry = Backoff{
	Steps:    defaultRetrySteps,
	Duration: defaultRetryDuration,
}

func OnError(ctx context.Context, backoff Backoff, fn func() error) error { //nolint:gocognit
	currentTry := 0

	for ctx.Err() == nil {
		err := fn()

		// if no error or error is not retriable, return
		if err == nil {
			return nil
		}

		// for cases when need to retry call to OpenAI API
		if errors.Is(err, ErrNotAnError) {
			continue
		}

		currentTry++

		if errors.Is(err, context.DeadlineExceeded) {
			// when context deadline exceeded, we should make retry as soon as possible
			if err = backoff.CanRetry(ctx, currentTry, err, time.Second); err != nil {
				return errors.Join(err, ErrMakingRetryableCall)
			}

			continue
		}

		// if error is retriable, try to retry
		if backoff.RetryableError != nil && backoff.RetryableError(err) {
			if err = backoff.CanRetry(ctx, currentTry, err, backoff.Duration); err != nil {
				return errors.Join(err, ErrMakingRetryableCall)
			}

			continue
		}

		// check if error is from OpenAI API
		apiError := &openai.APIError{}
		if errors.As(err, &apiError) {
			switch apiError.HTTPStatusCode {
			case http.StatusUnauthorized:
				return errors.Join(err, ErrInvalidAuth)
			case http.StatusTooManyRequests, http.StatusInternalServerError:
				if err = backoff.CanRetry(ctx, currentTry, err, backoff.Duration); err != nil {
					return errors.Join(err, ErrMakingRetryableCall)
				}

				continue
			default:
				return errors.Join(err, ErrMakingRetryableCall)
			}
		}

		// some error from other API
		return err
	}

	return errors.Join(ctx.Err(), ErrContextError)
}
