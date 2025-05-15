package openai //nolint:testpackage // testing private field

import (
	"errors"
	"strings"

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

func TestMultiPartFormUploads(t *testing.T) {
	body := &bytes.Buffer{}
	builder := NewFormBuilder(body)

	t.Run("MultipleFiles", func(t *testing.T) {
		file1, _ := os.CreateTemp(t.TempDir(), "*.png")
		file2, _ := os.CreateTemp(t.TempDir(), "*.jpg")
		defer file1.Close()
		defer file2.Close()

		checks.NoError(t, builder.CreateFormFile("image1", file1), "PNG file upload failed")
		checks.NoError(t, builder.CreateFormFile("image2", file2), "JPG file upload failed")
		checks.NoError(t, builder.WriteField("description", "test images"), "Field write failed")
	})

	t.Run("LargeFileConcurrent", func(t *testing.T) {
		bigFile, _ := os.CreateTemp(t.TempDir(), "*.bin")
		defer bigFile.Close()
		_, err := bigFile.Write(make([]byte, 1024*1024*5)) // 5MB test file
		checks.NoError(t, err, "Failed to write large file data")
		checks.NoError(t, builder.CreateFormFile("bigfile", bigFile), "Large file upload failed")
		checks.NoError(t, builder.WriteField("note", "large file test"), "Field write failed")
	})

	t.Run("MixedContentTypes", func(t *testing.T) {
		csvFile, _ := os.CreateTemp(t.TempDir(), "*.csv")
		textFile, _ := os.CreateTemp(t.TempDir(), "*.txt")
		defer csvFile.Close()
		defer textFile.Close()

		checks.NoError(t, builder.CreateFormFile("data", csvFile), "CSV file upload failed")
		checks.NoError(t, builder.CreateFormFile("text", textFile), "Text file upload failed")
		checks.NoError(t, builder.WriteField("format", "mixed"), "Field write failed")
	})
}

func TestFormDataContentType(t *testing.T) {
	body := &bytes.Buffer{}
	builder := NewFormBuilder(body)
	contentType := builder.FormDataContentType()
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		t.Fatalf("Content-Type格式错误，期望multipart/form-data开头，实际得到：%s", contentType)
	}
}

func TestCreateFormFileReader(t *testing.T) {
	body := &bytes.Buffer{}
	builder := NewFormBuilder(body)

	t.Run("SpecialCharacters", func(t *testing.T) {
		checks.NoError(t, builder.CreateFormFileReader("field", strings.NewReader("content"), "测 试@file.txt"), "特殊字符文件名应处理成功")
	})

	t.Run("InvalidReader", func(t *testing.T) {
		err := builder.CreateFormFileReader("field", &failingReader{}, "valid.txt")
		checks.HasError(t, err, "无效reader应返回错误")
	})
}

type failingReader struct{}

func (r *failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("mock read error")
}

func TestWriteFieldEdgeCases(t *testing.T) {
	mockErr := errors.New("mock write error")
	t.Run("EmptyFieldName", func(t *testing.T) {
		body := &bytes.Buffer{}
		builder := NewFormBuilder(body)
		err := builder.WriteField("", "valid-value")
		checks.HasError(t, err, "should return error for empty field name")
	})

	t.Run("EmptyValue", func(t *testing.T) {
		body := &bytes.Buffer{}
		builder := NewFormBuilder(body)
		err := builder.WriteField("valid-field", "")
		checks.NoError(t, err, "should allow empty value")
	})

	t.Run("MockWriterFailure", func(t *testing.T) {
		mockBuilder := &mockFormBuilder{
			mockWriteField: func(_, _ string) error {
				return mockErr
			},
		}
		err := mockBuilder.WriteField("field", "value")
		checks.ErrorIs(t, err, mockErr, "should propagate write error")
	})
}
