package openai_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestGPTImage26ModelSelection(t *testing.T) {
	for _, tt := range []struct {
		model string
		want  string
	}{
		{openai.CreateImageModelGptImage2Dot6Sunburst, "gpt-image-2.6-sunburst"},
		{openai.CreateImageModelGptImage2Dot6Flare, "gpt-image-2.6-flare"},
		{openai.CreateImageModelGptImage1Mini, "gpt-image-1-mini"},
	} {
		t.Run(tt.want, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %q; want POST", r.Method)
				}
				if model := imageModelFromRequest(t, r); model != tt.want {
					t.Errorf("%s model = %q; want %q", r.URL.Path, model, tt.want)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":[{"b64_json":"YQ=="}]}`)
			}))
			t.Cleanup(srv.Close)
			config := openai.DefaultConfig("fake-token")
			config.BaseURL = srv.URL
			client := openai.NewClientWithConfig(config)
			ctx := context.Background()
			created, err := client.CreateImage(ctx, openai.ImageRequest{Model: tt.model, Prompt: "A red square"})
			if err != nil {
				t.Fatalf("CreateImage: %v", err)
			}
			edited, err := client.CreateEditImage(ctx, openai.ImageEditRequest{
				Model: tt.model, Prompt: "Make the square blue",
				Image: openai.WrapReader(strings.NewReader("mock data"), "square.png", "image/png"),
			})
			if err != nil {
				t.Fatalf("CreateEditImage: %v", err)
			}
			checkGPTImageResponse(t, created)
			checkGPTImageResponse(t, edited)
		})
	}
}

func imageModelFromRequest(t *testing.T, r *http.Request) string {
	t.Helper()
	switch r.URL.Path {
	case "/images/generations":
		var request openai.ImageRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode generation request: %v", err)
		}
		return request.Model
	case "/images/edits":
		if err := r.ParseMultipartForm(4096); err != nil {
			t.Errorf("decode edit request: %v", err)
		}
		return r.FormValue("model")
	default:
		t.Errorf("unexpected request path: %s", r.URL.Path)
		return ""
	}
}

func checkGPTImageResponse(t *testing.T, got openai.ImageResponse) {
	t.Helper()
	if len(got.Data) != 1 || got.Data[0].B64JSON != "YQ==" {
		t.Errorf("unexpected response data: %+v", got.Data)
	}
}
