package openai_test

import (
	"bytes"
	"context"
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

	dir, cleanup := test.CreateTestDirectory(t)
	defer cleanup()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, "fake.mp3")
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

	dir, cleanup := test.CreateTestDirectory(t)
	defer cleanup()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, "fake.mp3")
			test.CreateTestFile(t, path)

			req := openai.AudioRequest{
				FilePath:    path,
				Model:       "whisper-3",
				Prompt:      "用简体中文",
				Temperature: 0.5,
				Language:    "zh",
				Format:      openai.AudioResponseFormatSRT,
			}
			_, err := tc.createFn(ctx, req)
			checks.NoError(t, err, "audio API error")
		})
	}
}

func TestAudioSpeechWithIncorrectParam(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	type args struct {
		request openai.SpeechRequest
		handler func(w http.ResponseWriter, r *http.Request)
	}
	tests := []struct {
		name      string
		wantErr   bool
		args      args
		wantBytes []byte
	}{
		{
			name: "NewSpeechRequest should return error on 500 status code",
			args: args{
				request: openai.NewSpeechRequest("test", openai.AudioSpeachModelTTS1, openai.AudioVoiceFable),
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					if _, err := w.Write([]byte(`Internal server error`)); err != nil {
						http.Error(w, "failed to write body", http.StatusInternalServerError)
					}
				},
			},
			wantErr:   true,
			wantBytes: nil,
		},
		{
			name: "NewSpeechRequest should return error on 404 status code",
			args: args{
				request: openai.NewSpeechRequest("test", openai.AudioSpeachModelTTS1, openai.AudioVoiceFable),
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					if _, err := w.Write([]byte(`{ "error": { "message": "Some error message"}}`)); err != nil {
						http.Error(w, "failed to write body", http.StatusInternalServerError)
					}
				},
			},
			wantErr:   true,
			wantBytes: nil,
		},
		{
			name: "NewSpeechRequest should return proper bytes",
			args: args{
				request: openai.NewSpeechRequest("test", openai.AudioSpeachModelTTS1, openai.AudioVoiceFable),
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte{0x01, 0x02, 0x03}); err != nil {
						http.Error(w, "failed to write body", http.StatusInternalServerError)
					}
				},
			},
			wantErr: false,
			wantBytes: []byte{
				0x01, 0x02, 0x03,
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		server.RegisterHandler("/v1/audio/speech", tt.args.handler)
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.CreateSpeechRaw(ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSpeechRaw error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if ok, err := bytesEqual(resp, tt.wantBytes); ok && err != nil {
					t.Errorf("CreateSpeechRaw return incorrect bytes")
					return
				}
			}
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

func bytesEqual(reader io.ReadCloser, bytes []byte) (ok bool, err error) {
	reads := make([]byte, 256)
	n, err := reader.Read(reads)
	if err != nil && err != io.EOF {
		return false, err
	}
	if len(bytes) != n {
		return false, err
	}
	for i := 0; i < n; i++ {
		if reads[i] != bytes[i] {
			return false, nil
		}
	}
	return true, nil
}
