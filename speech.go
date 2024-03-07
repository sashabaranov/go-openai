package openai

import (
	"context"
	"errors"
	"io"
	"net/http"
)

type SpeechModel string

const (
	TTSModel1      SpeechModel = "tts-1"
	TTSModel1HD    SpeechModel = "tts-1-hd"
	TTSModelCanary SpeechModel = "canary-tts"
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
	SpeechResponseFormatWav  SpeechResponseFormat = "wav"
	SpeechResponseFormatPcm  SpeechResponseFormat = "pcm"
)

var (
	ErrInvalidSpeechModel = errors.New("invalid speech model")
	ErrInvalidVoice       = errors.New("invalid voice")
)

type CreateSpeechRequest struct {
	Model          SpeechModel          `json:"model"`
	Input          string               `json:"input"`
	Voice          SpeechVoice          `json:"voice"`
	ResponseFormat SpeechResponseFormat `json:"response_format,omitempty"` // Optional, default to mp3
	Speed          float64              `json:"speed,omitempty"`           // Optional, default to 1.0
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func isValidSpeechModel(model SpeechModel) bool {
	return contains([]SpeechModel{TTSModel1, TTSModel1HD, TTSModelCanary}, model)
}

func isValidVoice(voice SpeechVoice) bool {
	return contains([]SpeechVoice{VoiceAlloy, VoiceEcho, VoiceFable, VoiceOnyx, VoiceNova, VoiceShimmer}, voice)
}

func (c *Client) CreateSpeech(ctx context.Context, request CreateSpeechRequest) (response io.ReadCloser, err error) {
	if !isValidSpeechModel(request.Model) {
		err = ErrInvalidSpeechModel
		return
	}
	if !isValidVoice(request.Voice) {
		err = ErrInvalidVoice
		return
	}
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/audio/speech", string(request.Model)),
		withBody(request),
		withContentType("application/json"),
	)
	if err != nil {
		return
	}

	response, err = c.sendRequestRaw(req)

	return
}
