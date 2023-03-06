package openai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// Whisper Defines the models provided by OpenAI to use when processing audio with OpenAI.
const (
	Whisper1 = "whisper-1"
)

// AudioRequest represents a request structure for audio API.
type AudioRequest struct {
	Model    string
	FilePath string
}

// AudioResponse represents a response structure for audio API.
type AudioResponse struct {
	Text string `json:"text"`
}

// CreateTranscription — API call to create a transcription. Returns transcribed text.
func (c *Client) CreateTranscription(
	ctx context.Context,
	request AudioRequest,
) (response AudioResponse, err error) {
	response, err = c.callAudioAPI(ctx, request, "transcriptions")
	return
}

// CreateTranslation — API call to translate audio into English.
func (c *Client) CreateTranslation(
	ctx context.Context,
	request AudioRequest,
) (response AudioResponse, err error) {
	response, err = c.callAudioAPI(ctx, request, "translations")
	return
}

// callAudioAPI — API call to an audio endpoint.
func (c *Client) callAudioAPI(
	ctx context.Context,
	request AudioRequest,
	endpointSuffix string,
) (response AudioResponse, err error) {
	var formBody bytes.Buffer
	w := multipart.NewWriter(&formBody)

	if err = audioMultipartForm(request, w); err != nil {
		return
	}

	urlSuffix := fmt.Sprintf("/audio/%s", endpointSuffix)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(urlSuffix), &formBody)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", w.FormDataContentType())

	err = c.sendRequest(req, &response)
	return
}

// audioMultipartForm creates a form with audio file contents and the name of the model to use for
// audio processing.
func audioMultipartForm(request AudioRequest, w *multipart.Writer) error {
	f, err := os.Open(request.FilePath)
	if err != nil {
		return fmt.Errorf("opening audio file: %w", err)
	}

	fw, err := w.CreateFormFile("file", f.Name())
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}

	if _, err = io.Copy(fw, f); err != nil {
		return fmt.Errorf("reading from opened audio file: %w", err)
	}

	fw, err = w.CreateFormField("model")
	if err != nil {
		return fmt.Errorf("creating form field: %w", err)
	}

	modelName := bytes.NewReader([]byte(request.Model))
	if _, err = io.Copy(fw, modelName); err != nil {
		return fmt.Errorf("writing model name: %w", err)
	}
	w.Close()

	return nil
}
