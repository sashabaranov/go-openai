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
// ResponseFormat is not supported for now. We only return JSON text, which may be sufficient.
type AudioRequest struct {
	Model       string
	FilePath    string
	Prompt      string // For translation, it should be in English
	Temperature float32
	Language    string // For translation, just do not use it. It seems "en" works, not confirmed...
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
	defer f.Close()

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

	// Create a form field for the prompt (if provided)
	if request.Prompt != "" {
		fw, err = w.CreateFormField("prompt")
		if err != nil {
			return fmt.Errorf("creating form field: %w", err)
		}

		prompt := bytes.NewReader([]byte(request.Prompt))
		if _, err = io.Copy(fw, prompt); err != nil {
			return fmt.Errorf("writing prompt: %w", err)
		}
	}

	// Create a form field for the temperature (if provided)
	if request.Temperature != 0 {
		fw, err = w.CreateFormField("temperature")
		if err != nil {
			return fmt.Errorf("creating form field: %w", err)
		}

		temperature := bytes.NewReader([]byte(fmt.Sprintf("%.2f", request.Temperature)))
		if _, err = io.Copy(fw, temperature); err != nil {
			return fmt.Errorf("writing temperature: %w", err)
		}
	}

	// Create a form field for the language (if provided)
	if request.Language != "" {
		fw, err = w.CreateFormField("language")
		if err != nil {
			return fmt.Errorf("creating form field: %w", err)
		}

		language := bytes.NewReader([]byte(request.Language))
		if _, err = io.Copy(fw, language); err != nil {
			return fmt.Errorf("writing language: %w", err)
		}
	}

	// Close the multipart writer
	w.Close()

	return nil
}
