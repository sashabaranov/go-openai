package openai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

// TestAudio Tests the transcription and translation endpoints of the API using the mocked server.
func TestAudio(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/audio/transcriptions", handleAudioEndpoint)
	server.RegisterHandler("/v1/audio/translations", handleAudioEndpoint)

	testcases := []struct {
		name     string
		createFn func(context.Context, openai.AudioRequest) (openai.AudioResponse, error)
	}{
		{
			"transcribe",
			client.CreateTranscription,
		},
		{
			"translate",
			client.CreateTranslation,
		},
	}

	ctx := context.Background()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "fake.mp3")
			test.CreateTestFile(t, path)

			req := openai.AudioRequest{
				FilePath: path,
				Model:    "whisper-3",
			}
			_, err := tc.createFn(ctx, req)
			checks.NoError(t, err, "audio API error")
		})

		t.Run(tc.name+" (with reader)", func(t *testing.T) {
			req := openai.AudioRequest{
				FilePath: "fake.webm",
				Reader:   bytes.NewBuffer([]byte(`some webm binary data`)),
				Model:    "whisper-3",
			}
			_, err := tc.createFn(ctx, req)
			checks.NoError(t, err, "audio API error")
		})
	}
}

func TestAudioWithOptionalArgs(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/audio/transcriptions", handleAudioEndpoint)
	server.RegisterHandler("/v1/audio/translations", handleAudioEndpoint)

	testcases := []struct {
		name     string
		createFn func(context.Context, openai.AudioRequest) (openai.AudioResponse, error)
	}{
		{
			"transcribe",
			client.CreateTranscription,
		},
		{
			"translate",
			client.CreateTranslation,
		},
	}

	ctx := context.Background()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "fake.mp3")
			test.CreateTestFile(t, path)

			req := openai.AudioRequest{
				FilePath:    path,
				Model:       "whisper-3",
				Prompt:      "用简体中文",
				Temperature: 0.5,
				Language:    "zh",
				Format:      openai.AudioResponseFormatSRT,
				TimestampGranularities: []openai.TranscriptionTimestampGranularity{
					openai.TranscriptionTimestampGranularitySegment,
					openai.TranscriptionTimestampGranularityWord,
				},
			}
			_, err := tc.createFn(ctx, req)
			checks.NoError(t, err, "audio API error")
		})
	}
}

// handleAudioEndpoint Handles the completion endpoint by the test server.
func handleAudioEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error

	// audio endpoints only accept POST requests
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "failed to parse media type", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(mediaType, "multipart") {
		http.Error(w, "request is not multipart", http.StatusBadRequest)
	}

	boundary, ok := params["boundary"]
	if !ok {
		http.Error(w, "no boundary in params", http.StatusBadRequest)
		return
	}

	fileData := &bytes.Buffer{}
	mr := multipart.NewReader(r.Body, boundary)
	part, err := mr.NextPart()
	if err != nil && errors.Is(err, io.EOF) {
		http.Error(w, "error accessing file", http.StatusBadRequest)
		return
	}
	if _, err = io.Copy(fileData, part); err != nil {
		http.Error(w, "failed to copy file", http.StatusInternalServerError)
		return
	}

	if len(fileData.Bytes()) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "received empty file data", http.StatusBadRequest)
		return
	}

	if _, err = w.Write([]byte(`{"body": "hello"}`)); err != nil {
		http.Error(w, "failed to write body", http.StatusInternalServerError)
		return
	}
}

func TestTranscriptionChunkingStrategy_JSONMarshal(t *testing.T) {
	t.Parallel()

	// Note: Form serialization in audioMultipartForm handles "auto" as literal string.
	// This test verifies standard JSON marshaling behavior.
	tests := []struct {
		name string
		cs   openai.TranscriptionChunkingStrategy
		want string
	}{
		{
			name: "auto serializes as object",
			cs:   openai.TranscriptionChunkingStrategy{Type: openai.AudioChunkingStrategyAuto},
			want: `{"type":"auto"}`,
		},
		{
			name: "server_vad with parameters",
			cs: openai.TranscriptionChunkingStrategy{
				Type:              openai.AudioChunkingStrategyServerVAD,
				PrefixPaddingMs:   300,
				SilenceDurationMs: 500,
				Threshold:         0.5,
			},
			want: `{"type":"server_vad","prefix_padding_ms":300,"silence_duration_ms":500,"threshold":0.5}`,
		},
		{
			name: "server_vad without optional params",
			cs:   openai.TranscriptionChunkingStrategy{Type: openai.AudioChunkingStrategyServerVAD},
			want: `{"type":"server_vad"}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := json.Marshal(tt.cs)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("json.Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAudioRequest_HasJSONResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format openai.AudioResponseFormat
		want   bool
	}{
		{"empty format defaults to JSON", "", true},
		{"json format", openai.AudioResponseFormatJSON, true},
		{"verbose_json format", openai.AudioResponseFormatVerboseJSON, true},
		{"diarized_json format", openai.AudioResponseFormatDiarizedJSON, true},
		{"text format", openai.AudioResponseFormatText, false},
		{"srt format", openai.AudioResponseFormatSRT, false},
		{"vtt format", openai.AudioResponseFormatVTT, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := openai.AudioRequest{Format: tt.format}
			if got := req.HasJSONResponse(); got != tt.want {
				t.Errorf("HasJSONResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateDiarizedTranscription(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/audio/transcriptions", handleDiarizedAudioEndpoint)

	ctx := context.Background()

	t.Run("with file path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "audio.mp3")
		test.CreateTestFile(t, path)

		resp, err := client.CreateDiarizedTranscription(ctx, openai.AudioRequest{
			FilePath: path,
			Model:    "gpt-4o-transcribe-diarize",
			Format:   openai.AudioResponseFormatDiarizedJSON,
		})
		checks.NoError(t, err, "CreateDiarizedTranscription error")

		if resp.Task != "transcribe" {
			t.Errorf("Task = %q, want %q", resp.Task, "transcribe")
		}
		if resp.Duration != 10.5 {
			t.Errorf("Duration = %v, want %v", resp.Duration, 10.5)
		}
		if len(resp.Segments) != 2 {
			t.Errorf("len(Segments) = %d, want %d", len(resp.Segments), 2)
		}
		if resp.Segments[0].Speaker != "A" {
			t.Errorf("Segments[0].Speaker = %q, want %q", resp.Segments[0].Speaker, "A")
		}
		if resp.Usage == nil {
			t.Error("Usage should not be nil")
		} else if resp.Usage.Seconds != 11 {
			t.Errorf("Usage.Seconds = %d, want %d", resp.Usage.Seconds, 11)
		}
	})

	t.Run("with reader", func(t *testing.T) {
		resp, err := client.CreateDiarizedTranscription(ctx, openai.AudioRequest{
			FilePath: "audio.mp3",
			Reader:   bytes.NewBuffer([]byte("audio data")),
			Model:    "gpt-4o-transcribe-diarize",
			Format:   openai.AudioResponseFormatDiarizedJSON,
		})
		checks.NoError(t, err, "CreateDiarizedTranscription error")

		if resp.Text == "" {
			t.Error("Text should not be empty")
		}
	})
}

func TestCreateDiarizedTranscription_WithChunkingStrategy(t *testing.T) {
	var receivedChunkingStrategy string
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/audio/transcriptions", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(1024 * 1024)
		if err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}
		receivedChunkingStrategy = r.FormValue("chunking_strategy")
		_, _ = w.Write([]byte(diarizedResponse))
	})

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "audio.mp3")
	test.CreateTestFile(t, path)

	t.Run("auto chunking", func(t *testing.T) {
		_, err := client.CreateDiarizedTranscription(ctx, openai.AudioRequest{
			FilePath: path,
			Model:    "gpt-4o-transcribe-diarize",
			Format:   openai.AudioResponseFormatDiarizedJSON,
			ChunkingStrategy: &openai.TranscriptionChunkingStrategy{
				Type: openai.AudioChunkingStrategyAuto,
			},
		})
		checks.NoError(t, err, "CreateDiarizedTranscription error")

		if receivedChunkingStrategy == "" {
			t.Fatal("chunking_strategy field was not sent")
		}

		// "auto" is sent as literal string, not JSON
		if receivedChunkingStrategy != "auto" {
			t.Errorf("chunking_strategy = %q, want %q", receivedChunkingStrategy, "auto")
		}
	})

	t.Run("server_vad with parameters", func(t *testing.T) {
		_, err := client.CreateDiarizedTranscription(ctx, openai.AudioRequest{
			FilePath: path,
			Model:    "gpt-4o-transcribe-diarize",
			Format:   openai.AudioResponseFormatDiarizedJSON,
			ChunkingStrategy: &openai.TranscriptionChunkingStrategy{
				Type:              openai.AudioChunkingStrategyServerVAD,
				PrefixPaddingMs:   300,
				SilenceDurationMs: 500,
				Threshold:         0.5,
			},
		})
		checks.NoError(t, err, "CreateDiarizedTranscription error")

		// "server_vad" serializes as a JSON object
		var cs openai.TranscriptionChunkingStrategy
		err = json.Unmarshal([]byte(receivedChunkingStrategy), &cs)
		if err != nil {
			t.Fatalf("failed to parse chunking_strategy: %v", err)
		}
		if cs.Type != openai.AudioChunkingStrategyServerVAD {
			t.Errorf("ChunkingStrategy.Type = %q, want %q", cs.Type, openai.AudioChunkingStrategyServerVAD)
		}
		if cs.PrefixPaddingMs != 300 {
			t.Errorf("ChunkingStrategy.PrefixPaddingMs = %d, want %d", cs.PrefixPaddingMs, 300)
		}
		if cs.SilenceDurationMs != 500 {
			t.Errorf("ChunkingStrategy.SilenceDurationMs = %d, want %d", cs.SilenceDurationMs, 500)
		}
		if cs.Threshold != 0.5 {
			t.Errorf("ChunkingStrategy.Threshold = %v, want %v", cs.Threshold, 0.5)
		}
	})

	t.Run("nil chunking strategy", func(t *testing.T) {
		receivedChunkingStrategy = "" // reset
		_, err := client.CreateDiarizedTranscription(ctx, openai.AudioRequest{
			FilePath:         path,
			Model:            "gpt-4o-transcribe-diarize",
			Format:           openai.AudioResponseFormatDiarizedJSON,
			ChunkingStrategy: nil,
		})
		checks.NoError(t, err, "CreateDiarizedTranscription error")

		if receivedChunkingStrategy != "" {
			t.Errorf("chunking_strategy should not be sent when nil, got %q", receivedChunkingStrategy)
		}
	})
}

func TestCreateDiarizedTranscription_FileNotFound(t *testing.T) {
	t.Parallel()

	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/audio/transcriptions", handleDiarizedAudioEndpoint)

	_, err := client.CreateDiarizedTranscription(context.Background(), openai.AudioRequest{
		FilePath: "/nonexistent/path/audio.mp3",
		Model:    "gpt-4o-transcribe-diarize",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "no such file") {
		t.Errorf("expected 'no such file' error, got: %v", err)
	}
}

func TestCreateDiarizedTranscription_HTTPError(t *testing.T) {
	t.Parallel()

	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/audio/transcriptions", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": {"message": "invalid api key", "type": "auth_error"}}`))
	})

	path := filepath.Join(t.TempDir(), "audio.mp3")
	test.CreateTestFile(t, path)

	_, err := client.CreateDiarizedTranscription(context.Background(), openai.AudioRequest{
		FilePath: path,
		Model:    "gpt-4o-transcribe-diarize",
	})
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

const diarizedResponse = `{
	"task": "transcribe",
	"duration": 10.5,
	"text": "[A]: Hello. [B]: Hi there.",
	"segments": [
		{
			"type": "transcript.text.segment",
			"id": "seg_001",
			"start": 0.0,
			"end": 1.5,
			"text": "Hello.",
			"speaker": "A"
		},
		{
			"type": "transcript.text.segment",
			"id": "seg_002",
			"start": 2.0,
			"end": 3.5,
			"text": "Hi there.",
			"speaker": "B"
		}
	],
	"usage": {
		"type": "duration",
		"seconds": 11
	}
}`

func handleDiarizedAudioEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "failed to parse media type", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(mediaType, "multipart") {
		http.Error(w, "request is not multipart", http.StatusBadRequest)
		return
	}

	boundary, ok := params["boundary"]
	if !ok {
		http.Error(w, "no boundary in params", http.StatusBadRequest)
		return
	}

	fileData := &bytes.Buffer{}
	mr := multipart.NewReader(r.Body, boundary)
	part, err := mr.NextPart()
	if err != nil && errors.Is(err, io.EOF) {
		http.Error(w, "error accessing file", http.StatusBadRequest)
		return
	}
	if _, err = io.Copy(fileData, part); err != nil {
		http.Error(w, "failed to copy file", http.StatusInternalServerError)
		return
	}

	if len(fileData.Bytes()) == 0 {
		http.Error(w, "received empty file data", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write([]byte(diarizedResponse)); err != nil {
		http.Error(w, "failed to write body", http.StatusInternalServerError)
		return
	}
}
