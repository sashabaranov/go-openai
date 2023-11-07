package openai

import (
	"context"
	"io"
	"net/http"
)

type SpeechModel string

const (
	TTSModel1   SpeechModel = "tts-1"
	TTsModel1HD SpeechModel = "tts-1-hd"
)

type SpeechVoice string

const (
	VoiceAlloy   SpeechVoice = "alloy"
	VoiceEcho    SpeechVoice = "echo"
	VoiceFable   SpeechVoice = "fable"
	VoiceOnyx    SpeechVoice = "onyx"
	VoiceNova    SpeechVoice = "nova"
	VoiceShimmer SpeechVoice = "shimmer"
)

type SpeechResponseFormat string

const (
	SpeechResponseFormatMp3  SpeechResponseFormat = "mp3"
	SpeechResponseFormatOpus SpeechResponseFormat = "opus"
	SpeechResponseFormatAac  SpeechResponseFormat = "aac"
	SpeechResponseFormatFlac SpeechResponseFormat = "flac"
)

type CreateSpeechRequest struct {
	Model          SpeechModel          `json:"model"`
	Input          string               `json:"input"`
	Voice          SpeechVoice          `json:"voice"`
	ResponseFormat SpeechResponseFormat `json:"response_format,omitempty"` // Optional, default to mp3
	Speed          float64              `json:"speed,omitempty"`           // Optional, default to 1.0
}

func (c *Client) CreateSpeech(ctx context.Context, request CreateSpeechRequest) (response io.ReadCloser, err error) {
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
