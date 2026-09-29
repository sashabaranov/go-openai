package openai_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestModernImageEditMultipart(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/images/edits", func(w http.ResponseWriter, r *http.Request) {
		checks.NoError(t, r.ParseMultipartForm(1<<20), "parse multipart form")
		defer func() { checks.NoError(t, r.MultipartForm.RemoveAll(), "remove multipart files") }()
		for key, want := range map[string]string{
			"model": "gpt-image-2.5-sunburst", "quality": "max", "prompt": "Make it blue",
			"size": "1536x1024", "background": "transparent", "output_format": "webp",
			"output_compression": "0", "user": "test-user", "n": "1",
		} {
			if got := r.FormValue(key); got != want {
				t.Errorf("%s = %q, want %q", key, got, want)
			}
		}
		if _, exists := r.MultipartForm.Value["response_format"]; exists {
			t.Error("unset response_format must be omitted for GPT Image")
		}
		file, header, err := r.FormFile("image")
		checks.NoError(t, err, "read image file")
		defer file.Close()
		if header.Filename != "input.png" || header.Header.Get("Content-Type") != "image/png" {
			t.Errorf("unexpected file metadata: %+v", header)
		}
		data, err := io.ReadAll(file)
		checks.NoError(t, err, "read image data")
		if string(data) != "image bytes" {
			t.Errorf("unexpected image data: %q", data)
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(w, `{"data":[{"b64_json":"aW1hZ2U="}]}`)
		checks.NoError(t, err, "write response")
	})
	compression := 0
	response, err := client.CreateEditImage(context.Background(), openai.ImageEditRequest{
		Image: openai.WrapReader(strings.NewReader("image bytes"), "input.png", "image/png"),
		Model: openai.CreateImageModelGptImage2Dot5Sunburst, Quality: openai.CreateImageQualityMax,
		Prompt: "Make it blue", Size: openai.CreateImageSize1536x1024, N: 1, User: "test-user",
		Background: openai.CreateImageBackgroundTransparent, OutputFormat: openai.CreateImageOutputFormatWEBP,
		OutputCompression: &compression,
	})
	checks.NoError(t, err, "edit image")
	if len(response.Data) != 1 || response.Data[0].B64JSON != "aW1hZ2U=" {
		t.Errorf("unexpected image output: %+v", response)
	}
}

func TestImageEditOmitsUnsetOptions(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/images/edits", func(w http.ResponseWriter, r *http.Request) {
		checks.NoError(t, r.ParseMultipartForm(1<<20), "parse multipart form")
		defer func() { checks.NoError(t, r.MultipartForm.RemoveAll(), "remove multipart files") }()
		if len(r.MultipartForm.Value) != 1 || r.FormValue("prompt") != "Edit" {
			t.Errorf("unset options must use API defaults: %+v", r.MultipartForm.Value)
		}
		w.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(w, `{"data":[]}`)
		checks.NoError(t, err, "write response")
	})
	_, err := client.CreateEditImage(context.Background(), openai.ImageEditRequest{
		Image: openai.WrapReader(strings.NewReader("image bytes"), "input.png", "image/png"), Prompt: "Edit",
	})
	checks.NoError(t, err, "edit image")
}
