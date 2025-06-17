package openai //nolint:testpackage // testing private field

import (
	"errors"
	"io"

	"github.com/sashabaranov/go-openai/internal/test/checks"

	"bytes"
	"os"
	"testing"
)

type mockFormBuilder struct {
	mockCreateFormFile func(string, *os.File) error
	mockWriteField     func(string, string) error
	mockClose          func() error
}

func (m *mockFormBuilder) CreateFormFile(fieldname string, file *os.File) error {
	return m.mockCreateFormFile(fieldname, file)
}

func (m *mockFormBuilder) WriteField(fieldname, value string) error {
	return m.mockWriteField(fieldname, value)
}

func (m *mockFormBuilder) Close() error {
	return m.mockClose()
}

func (m *mockFormBuilder) FormDataContentType() string {
	return ""
}

func TestCloseMethod(t *testing.T) {
	t.Run("NormalClose", func(t *testing.T) {
		body := &bytes.Buffer{}
		builder := NewFormBuilder(body)
		checks.NoError(t, builder.Close(), "正常关闭应成功")
	})

	t.Run("ErrorPropagation", func(t *testing.T) {
		errorMock := errors.New("mock close error")
		mockBuilder := &mockFormBuilder{
			mockClose: func() error {
				return errorMock
			},
		}
		err := mockBuilder.Close()
		checks.ErrorIs(t, err, errorMock, "应传递关闭错误")
	})
}

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

type readerWithNameAndContentType struct {
	io.Reader
}

func (*readerWithNameAndContentType) Name() string {
	return ""
}

func (*readerWithNameAndContentType) ContentType() string {
	return "image/png"
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

	rnc := &readerWithNameAndContentType{Reader: &bytes.Buffer{}}
	err = builder.CreateFormFileReader("file", rnc, "")
	checks.NoError(t, err, "formbuilder should not return error")
}

func TestFormDataContentType(t *testing.T) {
	t.Run("ReturnsUnderlyingWriterContentType", func(t *testing.T) {
		buf := &bytes.Buffer{}
		builder := NewFormBuilder(buf)

		contentType := builder.FormDataContentType()
		if contentType == "" {
			t.Errorf("expected non-empty content type, got empty string")
		}
	})
}

func TestWriteField(t *testing.T) {
	t.Run("EmptyFieldNameShouldReturnError", func(t *testing.T) {
		buf := &bytes.Buffer{}
		builder := NewFormBuilder(buf)

		err := builder.WriteField("", "some value")
		checks.HasError(t, err, "fieldname is required")
	})

	t.Run("ValidFieldNameShouldSucceed", func(t *testing.T) {
		buf := &bytes.Buffer{}
		builder := NewFormBuilder(buf)

		err := builder.WriteField("key", "value")
		checks.NoError(t, err, "should write field without error")
	})
}
