package openai_test

import (
	"testing"

	openai "github.com/sashabaranov/go-openai/internal"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestDefaultErrorAccumulator_WriteMultiple(t *testing.T) {
	ea, ok := openai.NewErrorAccumulator().(*openai.DefaultErrorAccumulator)
	if !ok {
		t.Fatal("type assertion to *DefaultErrorAccumulator failed")
	}
	checks.NoError(t, ea.Write([]byte("{\"error\": \"test1\"}")))
	checks.NoError(t, ea.Write([]byte("{\"error\": \"test2\"}")))

	expected := "{\"error\": \"test1\"}{\"error\": \"test2\"}"
	if string(ea.Bytes()) != expected {
		t.Fatalf("Expected %q, got %q", expected, ea.Bytes())
	}
}

func TestDefaultErrorAccumulator_EmptyBuffer(t *testing.T) {
	ea, ok := openai.NewErrorAccumulator().(*openai.DefaultErrorAccumulator)
	if !ok {
		t.Fatal("type assertion to *DefaultErrorAccumulator failed")
	}
	if len(ea.Bytes()) != 0 {
		t.Fatal("Buffer should be empty initially")
	}
}

func TestDefaultErrorAccumulator_WriteError(t *testing.T) {
	ea := &openai.DefaultErrorAccumulator{Buffer: &test.FailingErrorBuffer{}}
	err := ea.Write([]byte("fail"))
	checks.ErrorIs(t, err, test.ErrTestErrorAccumulatorWriteFailed, "Write should propagate buffer errors")
}
