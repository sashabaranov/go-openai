//nolint:testpackage // need access to unexported fields and types for testing
package test

import (
	"errors"
	"testing"
)

func TestFailingErrorBuffer(t *testing.T) {
	buf := &FailingErrorBuffer{}
	n, err := buf.Write([]byte("test"))
	if !errors.Is(err, ErrTestErrorAccumulatorWriteFailed) {
		t.Fatalf("expected %v, got %v", ErrTestErrorAccumulatorWriteFailed, err)
	}
	if n != 0 {
		t.Fatalf("expected n=0, got %d", n)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected Len 0, got %d", buf.Len())
	}
	if len(buf.Bytes()) != 0 {
		t.Fatalf("expected empty bytes")
	}
}
