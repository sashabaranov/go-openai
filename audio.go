package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	utils "github.com/sashabaranov/go-openai/internal"
)

// Whisper Defines the models provided by OpenAI to use when processing audio with OpenAI.
const (
	Whisper1 = "whisper-1"
)

// Response formats; Whisper uses AudioResponseFormatJSON by default.
type AudioResponseFormat string

const (
	AudioResponseFormatJSON         AudioResponseFormat = "json"
	AudioResponseFormatText         AudioResponseFormat = "text"
	AudioResponseFormatSRT          AudioResponseFormat = "srt"
	AudioResponseFormatVerboseJSON  AudioResponseFormat = "verbose_json"
	AudioResponseFormatVTT          AudioResponseFormat = "vtt"
	AudioResponseFormatDiarizedJSON AudioResponseFormat = "diarized_json"
)

type TranscriptionTimestampGranularity string

const (
	TranscriptionTimestampGranularityWord    TranscriptionTimestampGranularity = "word"
	TranscriptionTimestampGranularitySegment TranscriptionTimestampGranularity = "segment"
)

// AudioChunkingStrategyType defines the chunking strategy for audio transcription.
type AudioChunkingStrategyType string

// Chunking strategy types for audio transcription.
const (
	AudioChunkingStrategyAuto      AudioChunkingStrategyType = "auto"       // Server normalizes loudness and uses VAD
	AudioChunkingStrategyServerVAD AudioChunkingStrategyType = "server_vad" // Custom VAD parameters
)

// TranscriptionChunkingStrategy controls how audio is cut into chunks.
// When Type is ChunkingStrategyAuto ("auto"), the form field contains the literal string "auto".
// When Type is ChunkingStrategyServerVAD ("server_vad"), the form field contains a JSON object with VAD parameters.
// Required for gpt-4o-transcribe-diarize model on audio longer than 30 seconds.
type TranscriptionChunkingStrategy struct {
	// Type is AudioChunkingStrategyAuto or AudioChunkingStrategyServerVAD.
	Type AudioChunkingStrategyType `json:"type"`
	// PrefixPaddingMs is padding before detected speech (ms).
	PrefixPaddingMs int `json:"prefix_padding_ms,omitempty"`
	// SilenceDurationMs is silence threshold for chunk boundaries (ms).
	SilenceDurationMs int `json:"silence_duration_ms,omitempty"`
	// Threshold is VAD detection sensitivity (0.0-1.0).
	Threshold float32 `json:"threshold,omitempty"`
}

// toFormValue returns the string representation for multipart form submission.
// "auto" is sent as literal string; "server_vad" is sent as JSON object.
func (s TranscriptionChunkingStrategy) toFormValue() string {
	if s.Type == AudioChunkingStrategyAuto {
		return string(AudioChunkingStrategyAuto)
	}
	data, _ := json.Marshal(s)
	return string(data)
}

// AudioRequest represents a request structure for audio API.
type AudioRequest struct {
	Model string

	// FilePath is either an existing file in your filesystem or a filename representing the contents of Reader.
	FilePath string

	// Reader is an optional io.Reader when you do not want to use an existing file.
	Reader io.Reader

	Prompt                 string
	Temperature            float32
	Language               string // Only for transcription.
	Format                 AudioResponseFormat
	TimestampGranularities []TranscriptionTimestampGranularity // Only for transcription.
	// ChunkingStrategy controls audio chunking. Required for diarization models on audio >30s.
	ChunkingStrategy *TranscriptionChunkingStrategy
}

// AudioResponse represents a response structure for audio API.
type AudioResponse struct {
	Task     string  `json:"task"`
	Language string  `json:"language"`
	Duration float64 `json:"duration"`
	Segments []struct {
		ID               int     `json:"id"`
		Seek             int     `json:"seek"`
		Start            float64 `json:"start"`
		End              float64 `json:"end"`
		Text             string  `json:"text"`
		Tokens           []int   `json:"tokens"`
		Temperature      float64 `json:"temperature"`
		AvgLogprob       float64 `json:"avg_logprob"`
		CompressionRatio float64 `json:"compression_ratio"`
		NoSpeechProb     float64 `json:"no_speech_prob"`
		Transient        bool    `json:"transient"`
	} `json:"segments"`
	Words []struct {
		Word  string  `json:"word"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	} `json:"words"`
	Text string `json:"text"`

	httpHeader
}

// AudioUsage represents usage statistics for audio API calls.
type AudioUsage struct {
	Type    string `json:"type"`              // "duration" or "tokens"
	Seconds int    `json:"seconds,omitempty"` // Duration in seconds (for duration-based billing)
}

// DiarizedSegment represents a speaker-annotated segment from diarized transcription.
type DiarizedSegment struct {
	Type    string  `json:"type"`    // "transcript.text.segment"
	ID      string  `json:"id"`      // Segment identifier (e.g., "seg_001")
	Start   float64 `json:"start"`   // Start time in seconds
	End     float64 `json:"end"`     // End time in seconds
	Text    string  `json:"text"`    // Transcript text for this segment
	Speaker string  `json:"speaker"` // Speaker label (e.g., "agent", "A")
}

// DiarizedAudioResponse represents a diarized transcription response.
// Returned when using gpt-4o-transcribe-diarize model with diarized_json format.
type DiarizedAudioResponse struct {
	Task     string            `json:"task"`     // "transcribe"
	Duration float64           `json:"duration"` // Audio duration in seconds
	Text     string            `json:"text"`     // Full transcript with speaker prefixes
	Segments []DiarizedSegment `json:"segments"` // Speaker-annotated segments
	Usage    *AudioUsage       `json:"usage,omitempty"`

	httpHeader
}

type audioTextResponse struct {
	Text string `json:"text"`

	httpHeader
}

func (r *audioTextResponse) ToAudioResponse() AudioResponse {
	return AudioResponse{
		Text:       r.Text,
		httpHeader: r.httpHeader,
	}
}

// CreateTranscription — API call to create a transcription. Returns transcribed text.
func (c *Client) CreateTranscription(
	ctx context.Context,
	request AudioRequest,
) (response AudioResponse, err error) {
	return c.callAudioAPI(ctx, request, "transcriptions")
}

// CreateDiarizedTranscription transcribes audio with speaker diarization.
// Use with gpt-4o-transcribe-diarize model and AudioResponseFormatDiarizedJSON format.
// Requires ChunkingStrategy for audio longer than 30 seconds.
func (c *Client) CreateDiarizedTranscription(
	ctx context.Context,
	request AudioRequest,
) (response DiarizedAudioResponse, err error) {
	var formBody bytes.Buffer
	builder := c.createFormBuilder(&formBody)

	if err = audioMultipartForm(request, builder); err != nil {
		return DiarizedAudioResponse{}, err
	}

	urlSuffix := "/audio/transcriptions"
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix, withModel(request.Model)),
		withBody(&formBody),
		withContentType(builder.FormDataContentType()),
	)
	if err != nil {
		return DiarizedAudioResponse{}, err
	}

	err = c.sendRequest(req, &response)
	if err != nil {
		return DiarizedAudioResponse{}, err
	}
	return
}

// CreateTranslation — API call to translate audio into English.
func (c *Client) CreateTranslation(
	ctx context.Context,
	request AudioRequest,
) (response AudioResponse, err error) {
	return c.callAudioAPI(ctx, request, "translations")
}

// callAudioAPI — API call to an audio endpoint.
func (c *Client) callAudioAPI(
	ctx context.Context,
	request AudioRequest,
	endpointSuffix string,
) (response AudioResponse, err error) {
	var formBody bytes.Buffer
	builder := c.createFormBuilder(&formBody)

	if err = audioMultipartForm(request, builder); err != nil {
		return AudioResponse{}, err
	}

	urlSuffix := fmt.Sprintf("/audio/%s", endpointSuffix)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix, withModel(request.Model)),
		withBody(&formBody),
		withContentType(builder.FormDataContentType()),
	)
	if err != nil {
		return AudioResponse{}, err
	}

	if request.HasJSONResponse() {
		err = c.sendRequest(req, &response)
	} else {
		var textResponse audioTextResponse
		err = c.sendRequest(req, &textResponse)
		response = textResponse.ToAudioResponse()
	}
	if err != nil {
		return AudioResponse{}, err
	}
	return
}

// HasJSONResponse returns true if the response format is JSON.
func (r AudioRequest) HasJSONResponse() bool {
	return r.Format == "" || r.Format == AudioResponseFormatJSON ||
		r.Format == AudioResponseFormatVerboseJSON || r.Format == AudioResponseFormatDiarizedJSON
}

// audioMultipartForm creates a form with audio file contents and the name of the model to use for
// audio processing.
func audioMultipartForm(request AudioRequest, b utils.FormBuilder) error {
	err := createFileField(request, b)
	if err != nil {
		return err
	}

	err = b.WriteField("model", request.Model)
	if err != nil {
		return fmt.Errorf("writing model name: %w", err)
	}

	// Create a form field for the prompt (if provided)
	if request.Prompt != "" {
		err = b.WriteField("prompt", request.Prompt)
		if err != nil {
			return fmt.Errorf("writing prompt: %w", err)
		}
	}

	// Create a form field for the format (if provided)
	if request.Format != "" {
		err = b.WriteField("response_format", string(request.Format))
		if err != nil {
			return fmt.Errorf("writing format: %w", err)
		}
	}

	// Create a form field for the temperature (if provided)
	if request.Temperature != 0 {
		err = b.WriteField("temperature", fmt.Sprintf("%.2f", request.Temperature))
		if err != nil {
			return fmt.Errorf("writing temperature: %w", err)
		}
	}

	// Create a form field for the language (if provided)
	if request.Language != "" {
		err = b.WriteField("language", request.Language)
		if err != nil {
			return fmt.Errorf("writing language: %w", err)
		}
	}

	if err = writeTimestampGranularities(request.TimestampGranularities, b); err != nil {
		return err
	}

	if err = writeChunkingStrategy(request.ChunkingStrategy, b); err != nil {
		return err
	}

	// Close the multipart writer
	return b.Close()
}

// writeTimestampGranularities writes the timestamp_granularities[] fields if provided.
func writeTimestampGranularities(granularities []TranscriptionTimestampGranularity, b utils.FormBuilder) error {
	for _, tg := range granularities {
		if err := b.WriteField("timestamp_granularities[]", string(tg)); err != nil {
			return fmt.Errorf("writing timestamp_granularities[]: %w", err)
		}
	}
	return nil
}

// writeChunkingStrategy writes the chunking_strategy field if provided.
func writeChunkingStrategy(cs *TranscriptionChunkingStrategy, b utils.FormBuilder) error {
	if cs == nil {
		return nil
	}
	if err := b.WriteField("chunking_strategy", cs.toFormValue()); err != nil {
		return fmt.Errorf("writing chunking_strategy: %w", err)
	}
	return nil
}

// createFileField creates the "file" form field from either an existing file or by using the reader.
func createFileField(request AudioRequest, b utils.FormBuilder) error {
	if request.Reader != nil {
		err := b.CreateFormFileReader("file", request.Reader, request.FilePath)
		if err != nil {
			return fmt.Errorf("creating form using reader: %w", err)
		}
		return nil
	}

	f, err := os.Open(request.FilePath)
	if err != nil {
		return fmt.Errorf("opening audio file: %w", err)
	}
	defer f.Close()

	err = b.CreateFormFile("file", f)
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}

	return nil
}
