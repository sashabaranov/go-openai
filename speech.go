package openai

import (
	"context"
	"io"
	"net/http"
)

type SpeechModel string

const (
	TTS_MODEL_1    SpeechModel = "tts-1"
	TTS_MODEL_1_HD SpeechModel = "tts-1-hd"
)

type SpeechVoice string

const (
	VOICE_ALLOY   SpeechVoice = "alloy"
	VOICE_ECHO    SpeechVoice = "echo"
	VOICE_FABLE   SpeechVoice = "fable"
	VOICE_ONYX    SpeechVoice = "onyx"
	VOICE_NOVA    SpeechVoice = "nova"
	VOICE_SHIMMER SpeechVoice = "shimmer"
)

type SpeechResponseFormat string

const (
	SPEECH_RESPONSE_FORMAT_MP3  SpeechResponseFormat = "mp3"
	SPEECH_RESPONSE_FORMAT_OPUS SpeechResponseFormat = "opus"
	SPEECH_RESPONSE_FORMAT_AAC  SpeechResponseFormat = "aac"
	SPEECH_RESPONSE_FORMAT_FLAC SpeechResponseFormat = "flac"
)

type CreateSpeechRequest struct {
	Model          SpeechModel           `json:"model"`
	Input          string                `json:"input"`
	Voice          SpeechVoice           `json:"voice"`
	ResponseFormat *SpeechResponseFormat `json:"response_format,omitempty"` // Optional, default to mp3
	Speed          *float64              `json:"speed,omitempty"`           // Optional, default to 1.0
}

func (c *Client) CreateSpeech(ctx context.Context, request CreateSpeechRequest) (response io.ReadCloser, err error) {
	if request.Speed == nil {
		defaultSpeed := float64(1.0)
		request.Speed = &defaultSpeed
	}
	if request.ResponseFormat == nil {
		defaultResponseFormat := SPEECH_RESPONSE_FORMAT_MP3
		request.ResponseFormat = &defaultResponseFormat
	}
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/audio/speech", request.Model),
		withBody(request),
		withContentType("application/json; charset=utf-8"),
	)
	if err != nil {
		return
	}

	response, err = c.sendRequestRaw(req)

	return
}
