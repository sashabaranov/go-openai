package openai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	utils "github.com/sashabaranov/go-openai/internal"
)

// Voice param for TTS API endpoint, represents one of supported voices
type AudioSpeechVoice string

const (
	AudioVoiceAlloy   AudioSpeechVoice = "alloy"
	AudioVoiceEcho    AudioSpeechVoice = "echo"
	AudioVoiceFable   AudioSpeechVoice = "fable"
	AudioVoiceOnyx    AudioSpeechVoice = "onyx"
	AudioVoiceNova    AudioSpeechVoice = "nova"
	AudioVoiceShimmer AudioSpeechVoice = "shimmer"
)

// TTS model 
type AudioSpeechModel string

const (
	AudioSpeachModelTTS1   AudioSpeechModel = "tts-1"
	AudioSpeachModelTTS1HD AudioSpeechModel = "tts-1-hd"
)

// TTS output audio format
type AudioSpeechResponseFormat string

const (
	AudioSpeachResponseMp3  AudioSpeechResponseFormat = "mp3"
	AudioSpeachResponseOpus AudioSpeechResponseFormat = "opus"
	AudioSpeachResponseAac  AudioSpeechResponseFormat = "aac"
	AudioSpeachResponseFlac AudioSpeechResponseFormat = "flac"
)

// SpeechRequest represents API request for TTS endpoint
// It is highly recommended to use NewSpeechRequest to create a request
type SpeechRequest struct {
	Model          AudioSpeechModel          `json:"model"`
	Prompt          string                    `json:"input"`
	Voice          AudioSpeechVoice          `json:"voice"`
	ResponseFormat AudioSpeechResponseFormat `json:"response_format"`
	Speed          float32                   `json:"speed"`
}

type speechRequestOption func (opts *SpeechRequest)

// WithSpeed allows to set up speech speed option. Should be in between 0.25 and 4.0 with default value of 1.0
func WithSpeed(speed float32) speechRequestOption {
	return func(opts *SpeechRequest) {
		opts.Speed = speed
	}
}

// WithResponseFormat allows to set up audio format, by default MP3 is used
func WithResponseFormat(format AudioSpeechResponseFormat) speechRequestOption {
	return func(opts *SpeechRequest) {
		opts.ResponseFormat = format
	}
}

// NewSpeechRequest creates SpeechRequest with predefined parameters
// text - text to convert to speach
// model - TTS model to use, only AudioSpeachModelTTS1 and AudioSpeachModelTTS1HD are currently supported by API
// voice - TTS voice to be used, one of AudioVoiceAlloy, AudioVoiceEcho, AudioVoiceFable, AudioVoiceOnyx, AudioVoiceNova or AudioVoiceShimmer AudioSpeechVoice currently suported by API
func NewSpeechRequest(text string, model AudioSpeechModel, voice AudioSpeechVoice, opts ...speechRequestOption) SpeechRequest {
	request := SpeechRequest{
		Prompt: text,
		Model: model,
		Voice: voice,
		Speed: 1.0,
		ResponseFormat: AudioSpeachResponseMp3,
	}
	for _, setter := range opts {
		setter(&request)
	}
	return request
}

// Whisper Defines the models provided by OpenAI to use when processing audio with OpenAI.
const (
	Whisper1 = "whisper-1"
)

// Response formats; Whisper uses AudioResponseFormatJSON by default.
type AudioResponseFormat string

const (
	AudioResponseFormatJSON        AudioResponseFormat = "json"
	AudioResponseFormatText        AudioResponseFormat = "text"
	AudioResponseFormatSRT         AudioResponseFormat = "srt"
	AudioResponseFormatVerboseJSON AudioResponseFormat = "verbose_json"
	AudioResponseFormatVTT         AudioResponseFormat = "vtt"
)

// AudioRequest represents a request structure for audio API.
// ResponseFormat is not supported for now. We only return JSON text, which may be sufficient.
type AudioRequest struct {
	Model string

	// FilePath is either an existing file in your filesystem or a filename representing the contents of Reader.
	FilePath string

	// Reader is an optional io.Reader when you do not want to use an existing file.
	Reader io.Reader

	Prompt      string // For translation, it should be in English
	Temperature float32
	Language    string // For translation, just do not use it. It seems "en" works, not confirmed...
	Format      AudioResponseFormat
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
	Text string `json:"text"`

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

// CreateSpeech - API call to create Text To Speach request. Returns speech audio stream with requestd format.
func (c *Client) CreateSpeechRaw(
	ctx context.Context,
	request SpeechRequest,
) (response io.ReadCloser, err error) {
	if ok, err := validateSpeed(request.Speed); !ok {
		return nil, err
	}
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL("/audio/speech"),
		withContentType("application/json"),
		withBody(request),
	)
	if err != nil {
		return nil, err
	}
	return c.sendRequestRaw(req)
}

// CreateTranscription — API call to create a transcription. Returns transcribed text.
func (c *Client) CreateTranscription(
	ctx context.Context,
	request AudioRequest,
) (response AudioResponse, err error) {
	return c.callAudioAPI(ctx, request, "transcriptions")
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
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix, request.Model),
		withBody(&formBody), withContentType(builder.FormDataContentType()))
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
	return r.Format == "" || r.Format == AudioResponseFormatJSON || r.Format == AudioResponseFormatVerboseJSON
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

	// Close the multipart writer
	return b.Close()
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

func validateSpeed(speed float32) (ok bool, err error) {
	if speed < 0.25 || speed > 4.0 {
		return false, errors.New("speed should be from 0.25f up to 4.0f")
	}
	return true, nil
}
