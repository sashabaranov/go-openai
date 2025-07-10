package openai //nolint:testpackage // testing private field

import (
	utils "github.com/sashabaranov/go-openai/internal"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
)

type mockFormBuilder struct {
	mockCreateFormFile       func(string, *os.File) error
	mockCreateFormFileReader func(string, io.Reader, string) error
	mockWriteField           func(string, string) error
	mockClose                func() error
}

func (fb *mockFormBuilder) CreateFormFile(fieldname string, file *os.File) error {
	return fb.mockCreateFormFile(fieldname, file)
}

func (fb *mockFormBuilder) CreateFormFileReader(fieldname string, r io.Reader, filename string) error {
	return fb.mockCreateFormFileReader(fieldname, r, filename)
}

func (fb *mockFormBuilder) WriteField(fieldname, value string) error {
	return fb.mockWriteField(fieldname, value)
}

func (fb *mockFormBuilder) Close() error {
	return fb.mockClose()
}

func (fb *mockFormBuilder) FormDataContentType() string {
	return ""
}

func TestImageFormBuilderFailures(t *testing.T) {
	config := DefaultConfig("")
	config.BaseURL = ""
	client := NewClientWithConfig(config)

	mockBuilder := &mockFormBuilder{}
	client.createFormBuilder = func(io.Writer) utils.FormBuilder {
		return mockBuilder
	}
	ctx := context.Background()

	req := ImageEditRequest{
		Mask: &os.File{},
	}

	mockFailedErr := fmt.Errorf("mock form builder fail")
	mockBuilder.mockCreateFormFileReader = func(string, io.Reader, string) error {
		return mockFailedErr
	}
	_, err := client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	mockBuilder.mockCreateFormFileReader = func(name string, _ io.Reader, _ string) error {
		if name == "mask" {
			return mockFailedErr
		}
		return nil
	}
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	mockBuilder.mockCreateFormFile = func(string, *os.File) error {
		return nil
	}

	var failForField string
	mockBuilder.mockWriteField = func(fieldname, _ string) error {
		if fieldname == failForField {
			return mockFailedErr
		}
		return nil
	}

	failForField = "prompt"
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	failForField = "n"
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	failForField = "size"
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	failForField = "response_format"
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	failForField = ""
	mockBuilder.mockClose = func() error {
		return mockFailedErr
	}
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")
}

func TestVariImageFormBuilderFailures(t *testing.T) {
	config := DefaultConfig("")
	config.BaseURL = ""
	client := NewClientWithConfig(config)

	mockBuilder := &mockFormBuilder{}
	client.createFormBuilder = func(io.Writer) utils.FormBuilder {
		return mockBuilder
	}
	ctx := context.Background()

	req := ImageVariRequest{}

	mockFailedErr := fmt.Errorf("mock form builder fail")
	mockBuilder.mockCreateFormFileReader = func(string, io.Reader, string) error {
		return mockFailedErr
	}
	_, err := client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateVariImage should return error if form builder fails")

	mockBuilder.mockCreateFormFileReader = func(string, io.Reader, string) error {
		return nil
	}

	var failForField string
	mockBuilder.mockWriteField = func(fieldname, _ string) error {
		if fieldname == failForField {
			return mockFailedErr
		}
		return nil
	}

	failForField = "n"
	_, err = client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateVariImage should return error if form builder fails")

	failForField = "size"
	_, err = client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateVariImage should return error if form builder fails")

	failForField = "response_format"
	_, err = client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateVariImage should return error if form builder fails")

	failForField = ""
	mockBuilder.mockClose = func() error {
		return mockFailedErr
	}
	_, err = client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")
}

type testNamedReader struct{ io.Reader }

func (testNamedReader) Name() string { return "named.txt" }

func TestWrapReader(t *testing.T) {
	r := bytes.NewBufferString("data")
	wrapped := WrapReader(r, "file.png", "image/png")
	f, ok := wrapped.(interface {
		Name() string
		ContentType() string
	})
	if !ok {
		t.Fatal("wrapped reader missing Name or ContentType")
	}
	if f.Name() != "file.png" {
		t.Fatalf("expected name file.png, got %s", f.Name())
	}
	if f.ContentType() != "image/png" {
		t.Fatalf("expected content type image/png, got %s", f.ContentType())
	}

	// test name from underlying reader
	nr := testNamedReader{Reader: bytes.NewBufferString("d")}
	wrapped = WrapReader(nr, "", "text/plain")
	f, ok = wrapped.(interface {
		Name() string
		ContentType() string
	})
	if !ok {
		t.Fatal("wrapped named reader missing Name or ContentType")
	}
	if f.Name() != "named.txt" {
		t.Fatalf("expected name named.txt, got %s", f.Name())
	}
	if f.ContentType() != "text/plain" {
		t.Fatalf("expected content type text/plain, got %s", f.ContentType())
	}

	// no name provided
	wrapped = WrapReader(bytes.NewBuffer(nil), "", "")
	f2, ok := wrapped.(interface{ Name() string })
	if !ok {
		t.Fatal("wrapped anonymous reader missing Name")
	}
	if f2.Name() != "" {
		t.Fatalf("expected empty name, got %s", f2.Name())
	}
}
