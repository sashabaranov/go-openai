package checks_test

import (
	"errors"
	"testing"

	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestChecksSuccessPaths(t *testing.T) {
	checks.NoError(t, nil)
	checks.NoErrorF(t, nil)
	checks.HasError(t, errors.New("err"))
	target := errors.New("x")
	checks.ErrorIs(t, target, target)
	checks.ErrorIsF(t, target, target, "msg")
	checks.ErrorIsNot(t, errors.New("y"), target)
	checks.ErrorIsNotf(t, errors.New("y"), target, "msg")
}
