package test

import (
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"os"
	"testing"
)

// CreateTestFile creates a fake file with "hello" as the content.
func CreateTestFile(t *testing.T, path string) {
	file, err := os.Create(path)
	checks.NoError(t, err, "failed to create file")

	if _, err = file.WriteString("hello"); err != nil {
		t.Fatalf("failed to write to file %v", err)
	}
	file.Close()
}

// CreateTestDirectory creates a temporary folder which will be deleted when cleanup is called.
func CreateTestDirectory(t *testing.T) (path string, cleanup func()) {
	t.Helper()

	path, err := os.MkdirTemp(os.TempDir(), "")
	checks.NoError(t, err)

	return path, func() { os.RemoveAll(path) }
}
