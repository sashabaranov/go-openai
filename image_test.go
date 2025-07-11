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
	ctx := context.Background()
	mockFailedErr := fmt.Errorf("mock form builder fail")

	newClient := func(fb *mockFormBuilder) *Client {
		cfg := DefaultConfig("")
		cfg.BaseURL = ""
		c := NewClientWithConfig(cfg)
		c.createFormBuilder = func(io.Writer) utils.FormBuilder { return fb }
		return c
	}

	tests := []struct {
		name  string
		setup func(*mockFormBuilder)
		req   ImageEditRequest
	}{
		{
			name: "image",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return mockFailedErr }
				fb.mockWriteField = func(string, string) error { return nil }
				fb.mockClose = func() error { return nil }
			},
			req: ImageEditRequest{Image: bytes.NewBuffer(nil), Mask: bytes.NewBuffer(nil)},
		},
		{
			name: "mask",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(name string, _ io.Reader, _ string) error {
					if name == "mask" {
						return mockFailedErr
					}
					return nil
				}
				fb.mockWriteField = func(string, string) error { return nil }
				fb.mockClose = func() error { return nil }
			},
			req: ImageEditRequest{Image: bytes.NewBuffer(nil), Mask: bytes.NewBuffer(nil)},
		},
		{
			name: "prompt",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return nil }
				fb.mockWriteField = func(field, _ string) error {
					if field == "prompt" {
						return mockFailedErr
					}
					return nil
				}
				fb.mockClose = func() error { return nil }
			},
			req: ImageEditRequest{Image: bytes.NewBuffer(nil), Mask: bytes.NewBuffer(nil)},
		},
		{
			name: "n",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return nil }
				fb.mockWriteField = func(field, _ string) error {
					if field == "n" {
						return mockFailedErr
					}
					return nil
				}
				fb.mockClose = func() error { return nil }
			},
			req: ImageEditRequest{Image: bytes.NewBuffer(nil), Mask: bytes.NewBuffer(nil)},
		},
		{
			name: "size",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return nil }
				fb.mockWriteField = func(field, _ string) error {
					if field == "size" {
						return mockFailedErr
					}
					return nil
				}
				fb.mockClose = func() error { return nil }
			},
			req: ImageEditRequest{Image: bytes.NewBuffer(nil), Mask: bytes.NewBuffer(nil)},
		},
		{
			name: "response_format",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return nil }
				fb.mockWriteField = func(field, _ string) error {
					if field == "response_format" {
						return mockFailedErr
					}
					return nil
				}
				fb.mockClose = func() error { return nil }
			},
			req: ImageEditRequest{Image: bytes.NewBuffer(nil), Mask: bytes.NewBuffer(nil)},
		},
		{
			name: "close",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return nil }
				fb.mockWriteField = func(string, string) error { return nil }
				fb.mockClose = func() error { return mockFailedErr }
			},
			req: ImageEditRequest{Image: bytes.NewBuffer(nil), Mask: bytes.NewBuffer(nil)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fb := &mockFormBuilder{}
			tc.setup(fb)
			client := newClient(fb)
			_, err := client.CreateEditImage(ctx, tc.req)
			checks.ErrorIs(t, err, mockFailedErr, "CreateEditImage should return error if form builder fails")
		})
	}

	t.Run("new request", func(t *testing.T) {
		fb := &mockFormBuilder{
			mockCreateFormFileReader: func(string, io.Reader, string) error { return nil },
			mockWriteField:           func(string, string) error { return nil },
			mockClose:                func() error { return nil },
		}
		client := newClient(fb)
		client.requestBuilder = &failingRequestBuilder{}

		_, err := client.CreateEditImage(ctx, ImageEditRequest{Image: bytes.NewBuffer(nil), Mask: bytes.NewBuffer(nil)})
		checks.ErrorIs(t, err, errTestRequestBuilderFailed, "CreateEditImage should return error if request builder fails")
	})
}

func TestVariImageFormBuilderFailures(t *testing.T) {
	ctx := context.Background()
	mockFailedErr := fmt.Errorf("mock form builder fail")

	newClient := func(fb *mockFormBuilder) *Client {
		cfg := DefaultConfig("")
		cfg.BaseURL = ""
		c := NewClientWithConfig(cfg)
		c.createFormBuilder = func(io.Writer) utils.FormBuilder { return fb }
		return c
	}

	tests := []struct {
		name  string
		setup func(*mockFormBuilder)
		req   ImageVariRequest
	}{
		{
			name: "image",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return mockFailedErr }
				fb.mockWriteField = func(string, string) error { return nil }
				fb.mockClose = func() error { return nil }
			},
			req: ImageVariRequest{Image: bytes.NewBuffer(nil)},
		},
		{
			name: "n",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return nil }
				fb.mockWriteField = func(field string, _ string) error {
					if field == "n" {
						return mockFailedErr
					}
					return nil
				}
				fb.mockClose = func() error { return nil }
			},
			req: ImageVariRequest{Image: bytes.NewBuffer(nil)},
		},
		{
			name: "size",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return nil }
				fb.mockWriteField = func(field string, _ string) error {
					if field == "size" {
						return mockFailedErr
					}
					return nil
				}
				fb.mockClose = func() error { return nil }
			},
			req: ImageVariRequest{Image: bytes.NewBuffer(nil)},
		},
		{
			name: "response_format",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return nil }
				fb.mockWriteField = func(field string, _ string) error {
					if field == "response_format" {
						return mockFailedErr
					}
					return nil
				}
				fb.mockClose = func() error { return nil }
			},
			req: ImageVariRequest{Image: bytes.NewBuffer(nil)},
		},
		{
			name: "close",
			setup: func(fb *mockFormBuilder) {
				fb.mockCreateFormFileReader = func(string, io.Reader, string) error { return nil }
				fb.mockWriteField = func(string, string) error { return nil }
				fb.mockClose = func() error { return mockFailedErr }
			},
			req: ImageVariRequest{Image: bytes.NewBuffer(nil)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fb := &mockFormBuilder{}
			tc.setup(fb)
			client := newClient(fb)
			_, err := client.CreateVariImage(ctx, tc.req)
			checks.ErrorIs(t, err, mockFailedErr, "CreateVariImage should return error if form builder fails")
		})
	}

	t.Run("new request", func(t *testing.T) {
		fb := &mockFormBuilder{
			mockCreateFormFileReader: func(string, io.Reader, string) error { return nil },
			mockWriteField:           func(string, string) error { return nil },
			mockClose:                func() error { return nil },
		}
		client := newClient(fb)
		client.requestBuilder = &failingRequestBuilder{}

		_, err := client.CreateVariImage(ctx, ImageVariRequest{Image: bytes.NewBuffer(nil)})
		checks.ErrorIs(t, err, errTestRequestBuilderFailed, "CreateVariImage should return error if request builder fails")
	})
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
