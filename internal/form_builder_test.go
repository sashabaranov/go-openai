package openai //nolint:testpackage // testing private field

import (
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"bytes"
	"errors"
	"os"
	"testing"
)

type failingWriter struct {
}

var errMockFailingWriterError = errors.New("mock writer failed")

func (*failingWriter) Write([]byte) (int, error) {
	return 0, errMockFailingWriterError
}

func TestFormBuilderWithFailingWriter(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "")
	if err != nil {
		t.Fatalf("Error creating tmp file: %v", err)
	}
	defer file.Close()

	builder := NewFormBuilder(&failingWriter{})
	err = builder.CreateFormFile("file", file)
	checks.ErrorIs(t, err, errMockFailingWriterError, "formbuilder should return error if writer fails")
}

func TestFormBuilderWithClosedFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "")
	if err != nil {
		t.Fatalf("Error creating tmp file: %v", err)
	}
	file.Close()

	body := &bytes.Buffer{}
	builder := NewFormBuilder(body)
	err = builder.CreateFormFile("file", file)
	checks.HasError(t, err, "formbuilder should return error if file is closed")
	checks.ErrorIs(t, err, os.ErrClosed, "formbuilder should return error if file is closed")
}

type failingReader struct {
}

var errMockFailingReaderError = errors.New("mock reader failed")

func (*failingReader) Read([]byte) (int, error) {
	return 0, errMockFailingReaderError
}

func TestFormBuilderWithReader(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "")
	if err != nil {
		t.Fatalf("Error creating tmp file: %v", err)
	}
	defer file.Close()
	builder := NewFormBuilder(&failingWriter{})
	err = builder.CreateFormFileReader("file", file, file.Name())
	checks.ErrorIs(t, err, errMockFailingWriterError, "formbuilder should return error if writer fails")

	builder = NewFormBuilder(&bytes.Buffer{})
	reader := &failingReader{}
	err = builder.CreateFormFileReader("file", reader, "")
	checks.ErrorIs(t, err, errMockFailingReaderError, "formbuilder should return error if copy reader fails")

	successReader := &bytes.Buffer{}
	err = builder.CreateFormFileReader("file", successReader, "")
	checks.NoError(t, err, "formbuilder should not return error")
}
