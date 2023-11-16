package openai //nolint:testpackage // testing private field

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	utils "github.com/sashabaranov/go-openai/internal"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestFileUploadWithFailingFormBuilder(t *testing.T) {
	config := DefaultConfig("")
	config.BaseURL = ""
	client := NewClientWithConfig(config)
	mockBuilder := &mockFormBuilder{}
	client.createFormBuilder = func(io.Writer) utils.FormBuilder {
		return mockBuilder
	}

	ctx := context.Background()
	req := FileRequest{
		FileName: "test.go",
		FilePath: "client.go",
		Purpose:  PurposeAssistants,
	}

	// mock WriteFiled with given error
	fieldWith := func(err error) func(string, string) error {
		return func(string, string) error {
			return err
		}
	}
	// mock CreateFormFileReader with given error
	ffrWith := func(err error) func(string, io.Reader, string) error {
		return func(string, io.Reader, string) error {
			return err
		}
	}
	mockError := fmt.Errorf("mockWriteField error")
	mockBuilder.mockWriteField = fieldWith(mockError)
	_, err := client.CreateFile(ctx, req)
	checks.ErrorIs(t, err, mockError, "CreateFile should return error if form builder fails")

	mockError = fmt.Errorf("mockCreateFormFile error")
	mockBuilder.mockWriteField = fieldWith(nil)
	mockBuilder.mockCreateFormFileReader = ffrWith(mockError)
	_, err = client.CreateFile(ctx, req)
	checks.ErrorIs(t, err, mockError, "CreateFile should return error if form builder fails")

	mockError = fmt.Errorf("mockClose error")
	mockBuilder.mockWriteField = fieldWith(nil)
	mockBuilder.mockCreateFormFileReader = ffrWith(nil)
	mockBuilder.mockClose = func() error {
		return mockError
	}
	_, err = client.CreateFile(ctx, req)
	checks.ErrorIs(t, err, mockError, "CreateFile should return error if form builder fails")
}

func TestFileUploadWithNonExistentPath(t *testing.T) {
	config := DefaultConfig("")
	config.BaseURL = ""
	client := NewClientWithConfig(config)

	ctx := context.Background()
	req := FileRequest{
		FilePath: "some non existent file path/F616FD18-589E-44A8-BF0C-891EAE69C455",
		Purpose:  PurposeAssistants,
	}

	_, err := client.CreateFile(ctx, req)
	checks.ErrorIs(t, err, os.ErrNotExist, "CreateFile should return error if file does not exist")
}
