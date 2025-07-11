package checks_test

import (
	"errors"
	"fmt"
	"testing"

	checks "github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestNoError(t *testing.T) {
	checks.NoError(t, nil)
	checks.NoErrorF(t, nil)
	checks.HasError(t, errors.New("err"))
}

func TestErrorComparisons(t *testing.T) {
	target := errors.New("target")
	wrapped := fmt.Errorf("wrap: %w", target)

	checks.ErrorIs(t, wrapped, target)
	checks.ErrorIsF(t, wrapped, target, "%v")
	checks.ErrorIsNot(t, errors.New("other"), target)
	checks.ErrorIsNotf(t, errors.New("other"), target, "%v")
}
