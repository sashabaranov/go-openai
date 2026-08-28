package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	utils "github.com/sashabaranov/go-openai/internal"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestAudioWithFailingFormBuilder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fake.mp3")
	test.CreateTestFile(t, path)

	req := AudioRequest{
		FilePath:    path,
		Prompt:      "test",
		Temperature: 0.5,
		Language:    "en",
		Format:      AudioResponseFormatSRT,
		TimestampGranularities: []TranscriptionTimestampGranularity{
			TranscriptionTimestampGranularitySegment,
			TranscriptionTimestampGranularityWord,
		},
		ChunkingStrategy: "auto",
	}

	mockFailedErr := fmt.Errorf("mock form builder fail")
	mockBuilder := &mockFormBuilder{}

	mockBuilder.mockCreateFormFile = func(string, *os.File) error {
		return mockFailedErr
	}
	err := audioMultipartForm(req, mockBuilder)
	checks.ErrorIs(t, err, mockFailedErr, "audioMultipartForm should return error if form builder fails")

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

	failOn := []string{
		"model", "prompt", "temperature", "language",
		"response_format", "timestamp_granularities[]", "chunking_strategy",
	}
	for _, failingField := range failOn {
		failForField = failingField
		mockFailedErr = fmt.Errorf("mock form builder fail on field %s", failingField)

		err = audioMultipartForm(req, mockBuilder)
		checks.ErrorIs(t, err, mockFailedErr, "audioMultipartForm should return error if form builder fails")
	}
}

// captureChunkingStrategy runs audioMultipartForm and returns the value written for the
// chunking_strategy field, or ("", false) if the field was never written.
func captureChunkingStrategy(t *testing.T, path string, strategy any) (string, bool) {
	t.Helper()

	var got string
	seen := false
	mockBuilder := &mockFormBuilder{
		mockCreateFormFile: func(string, *os.File) error { return nil },
		mockWriteField: func(field, value string) error {
			if field == "chunking_strategy" {
				seen = true
				got = value
			}
			return nil
		},
		mockClose: func() error { return nil },
	}

	req := AudioRequest{FilePath: path, Model: GPT4oTranscribeDiarize, ChunkingStrategy: strategy}
	if err := audioMultipartForm(req, mockBuilder); err != nil {
		t.Fatalf("audioMultipartForm returned error: %v", err)
	}
	return got, seen
}

func TestAudioChunkingStrategyField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fake.mp3")
	test.CreateTestFile(t, path)

	vadStrategy := TranscriptionChunkingStrategy{
		Type:              "server_vad",
		PrefixPaddingMs:   300,
		SilenceDurationMs: 200,
		Threshold:         0.5,
	}
	vadJSON := `{"type":"server_vad","prefix_padding_ms":300,"silence_duration_ms":200,"threshold":0.5}`

	tests := []struct {
		name     string
		strategy any
		want     string
	}{
		{name: "auto string", strategy: "auto", want: "auto"},
		{name: "server_vad struct", strategy: vadStrategy, want: vadJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, seen := captureChunkingStrategy(t, path, tt.strategy)
			if !seen {
				t.Fatalf("chunking_strategy field was not written")
			}
			if got != tt.want {
				t.Fatalf("chunking_strategy = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("unset", func(t *testing.T) {
		if _, seen := captureChunkingStrategy(t, path, nil); seen {
			t.Fatalf("chunking_strategy should not be written when unset")
		}
	})
}

func TestCreateFileField(t *testing.T) {
	t.Run("createFileField failing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "fake.mp3")
		test.CreateTestFile(t, path)

		req := AudioRequest{
			FilePath: path,
		}

		mockFailedErr := fmt.Errorf("mock form builder fail")
		mockBuilder := &mockFormBuilder{
			mockCreateFormFile: func(string, *os.File) error {
				return mockFailedErr
			},
		}

		err := createFileField(req, mockBuilder)
		checks.ErrorIs(t, err, mockFailedErr, "createFileField using a file should return error if form builder fails")
	})

	t.Run("createFileField failing reader", func(t *testing.T) {
		req := AudioRequest{
			FilePath: "test.wav",
			Reader:   bytes.NewBuffer([]byte(`wav test contents`)),
		}

		mockFailedErr := fmt.Errorf("mock form builder fail")
		mockBuilder := &mockFormBuilder{
			mockCreateFormFileReader: func(string, io.Reader, string) error {
				return mockFailedErr
			},
		}

		err := createFileField(req, mockBuilder)
		checks.ErrorIs(t, err, mockFailedErr, "createFileField using a reader should return error if form builder fails")
	})

	t.Run("createFileField failing open", func(t *testing.T) {
		req := AudioRequest{
			FilePath: "non_existing_file.wav",
		}

		mockBuilder := &mockFormBuilder{}

		err := createFileField(req, mockBuilder)
		checks.HasError(t, err, "createFileField using file should return error when open file fails")
	})
}

// failingFormBuilder always returns an error when creating form files.
type failingFormBuilder struct{ err error }

func (f *failingFormBuilder) CreateFormFile(_ string, _ *os.File) error {
	return f.err
}

func (f *failingFormBuilder) CreateFormFileReader(_ string, _ io.Reader, _ string) error {
	return f.err
}

func (f *failingFormBuilder) WriteField(_, _ string) error {
	return nil
}

func (f *failingFormBuilder) Close() error {
	return nil
}

func (f *failingFormBuilder) FormDataContentType() string {
	return "multipart/form-data"
}

// failingAudioRequestBuilder simulates an error during HTTP request construction.
type failingAudioRequestBuilder struct{ err error }

func (f *failingAudioRequestBuilder) Build(
	_ context.Context,
	_, _ string,
	_ any,
	_ http.Header,
) (*http.Request, error) {
	return nil, f.err
}

// errorHTTPClient always returns an error when making HTTP calls.
type errorHTTPClient struct{ err error }

func (e *errorHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return nil, e.err
}

func TestCallAudioAPIMultipartFormError(t *testing.T) {
	client := NewClient("test-token")
	errForm := errors.New("mock create form file failure")
	// Override form builder to force an error during multipart form creation.
	client.createFormBuilder = func(_ io.Writer) utils.FormBuilder {
		return &failingFormBuilder{err: errForm}
	}

	// Provide a reader so createFileField uses the reader path (no file open).
	req := AudioRequest{FilePath: "fake.mp3", Reader: bytes.NewBuffer([]byte("dummy")), Model: Whisper1}
	_, err := client.callAudioAPI(context.Background(), req, "transcriptions")
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if !errors.Is(err, errForm) {
		t.Errorf("expected error %v, got %v", errForm, err)
	}
}

func TestCallAudioAPINewRequestError(t *testing.T) {
	client := NewClient("test-token")
	// Create a real temp file so multipart form succeeds.
	tmp := t.TempDir()
	path := filepath.Join(tmp, "file.mp3")
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	errBuild := errors.New("mock build failure")
	client.requestBuilder = &failingAudioRequestBuilder{err: errBuild}

	req := AudioRequest{FilePath: path, Model: Whisper1}
	_, err := client.callAudioAPI(context.Background(), req, "translations")
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if !errors.Is(err, errBuild) {
		t.Errorf("expected error %v, got %v", errBuild, err)
	}
}

func TestCallAudioAPISendRequestErrorJSON(t *testing.T) {
	client := NewClient("test-token")
	// Create a real temp file so multipart form succeeds.
	tmp := t.TempDir()
	path := filepath.Join(tmp, "file.mp3")
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	errHTTP := errors.New("mock HTTPClient failure")
	// Override HTTP client to simulate a network error.
	client.config.HTTPClient = &errorHTTPClient{err: errHTTP}

	req := AudioRequest{FilePath: path, Model: Whisper1}
	_, err := client.callAudioAPI(context.Background(), req, "transcriptions")
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if !errors.Is(err, errHTTP) {
		t.Errorf("expected error %v, got %v", errHTTP, err)
	}
}

func TestCallAudioAPISendRequestErrorText(t *testing.T) {
	client := NewClient("test-token")
	tmp := t.TempDir()
	path := filepath.Join(tmp, "file.mp3")
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	errHTTP := errors.New("mock HTTPClient failure")
	client.config.HTTPClient = &errorHTTPClient{err: errHTTP}

	// Use a non-JSON response format to exercise the text path.
	req := AudioRequest{FilePath: path, Model: Whisper1, Format: AudioResponseFormatText}
	_, err := client.callAudioAPI(context.Background(), req, "translations")
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if !errors.Is(err, errHTTP) {
		t.Errorf("expected error %v, got %v", errHTTP, err)
	}
}
