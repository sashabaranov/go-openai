package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestImages(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/images/generations", handleImageEndpoint)
	_, err := client.CreateImage(context.Background(), openai.ImageRequest{
		Prompt:         "Lorem ipsum",
		Model:          openai.CreateImageModelDallE3,
		N:              1,
		Quality:        openai.CreateImageQualityHD,
		Size:           openai.CreateImageSize1024x1024,
		Style:          openai.CreateImageStyleVivid,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		User:           "user",
	})
	checks.NoError(t, err, "CreateImage error")
}

// handleImageEndpoint Handles the images endpoint by the test server.
func handleImageEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// images only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	var imageReq openai.ImageRequest
	if imageReq, err = getImageBody(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}
	res := openai.ImageResponse{
		Created: time.Now().Unix(),
	}
	for i := 0; i < imageReq.N; i++ {
		imageData := openai.ImageResponseDataInner{}
		switch imageReq.ResponseFormat {
		case openai.CreateImageResponseFormatURL, "":
			imageData.URL = "https://example.com/image.png"
		case openai.CreateImageResponseFormatB64JSON:
			// This decodes to "{}" in base64.
			imageData.B64JSON = "e30K"
		default:
			http.Error(w, "invalid response format", http.StatusBadRequest)
			return
		}
		res.Data = append(res.Data, imageData)
	}
	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}

// getImageBody Returns the body of the request to create a image.
func getImageBody(r *http.Request) (openai.ImageRequest, error) {
	image := openai.ImageRequest{}
	// read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return openai.ImageRequest{}, err
	}
	err = json.Unmarshal(reqBody, &image)
	if err != nil {
		return openai.ImageRequest{}, err
	}
	return image, nil
}

func TestImageEdit(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/images/edits", handleEditImageEndpoint)

	origin, err := os.Create(filepath.Join(t.TempDir(), "image.png"))
	if err != nil {
		t.Fatalf("open origin file error: %v", err)
	}
	defer origin.Close()

	mask, err := os.Create(filepath.Join(t.TempDir(), "mask.png"))
	if err != nil {
		t.Fatalf("open mask file error: %v", err)
	}
	defer mask.Close()

	_, err = client.CreateEditImage(context.Background(), openai.ImageEditRequest{
		Image:          origin,
		Mask:           mask,
		Prompt:         "There is a turtle in the pool",
		N:              3,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
	})
	checks.NoError(t, err, "CreateImage error")
}

func TestImageEditWithoutMask(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/images/edits", handleEditImageEndpoint)

	origin, err := os.Create(filepath.Join(t.TempDir(), "image.png"))
	if err != nil {
		t.Fatalf("open origin file error: %v", err)
	}
	defer origin.Close()

	_, err = client.CreateEditImage(context.Background(), openai.ImageEditRequest{
		Image:          origin,
		Prompt:         "There is a turtle in the pool",
		N:              3,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
	})
	checks.NoError(t, err, "CreateImage error")
}

func TestImageEditSendsOptionalFields(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/images/edits", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm error: %v", err)
			http.Error(w, "could not parse multipart form", http.StatusBadRequest)
			return
		}

		for field, want := range map[string]string{
			"model":   openai.CreateImageModelGptImage1,
			"quality": openai.CreateImageQualityHigh,
			"user":    "test-user",
		} {
			if got := r.FormValue(field); got != want {
				t.Errorf("%s = %q, want %q", field, got, want)
			}
		}

		_, err := fmt.Fprintln(w, `{}`)
		checks.NoError(t, err, "Write response error")
	})

	_, err := client.CreateEditImage(context.Background(), openai.ImageEditRequest{
		Image:   openai.WrapReader(strings.NewReader("image"), "image.png", "image/png"),
		Prompt:  "There is a turtle in the pool",
		Model:   openai.CreateImageModelGptImage1,
		Quality: openai.CreateImageQualityHigh,
		User:    "test-user",
	})
	checks.NoError(t, err, "CreateEditImage error")
}

// handleEditImageEndpoint Handles the images endpoint by the test server.
func handleEditImageEndpoint(w http.ResponseWriter, r *http.Request) {
	var resBytes []byte

	// images only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	responses := openai.ImageResponse{
		Created: time.Now().Unix(),
		Data: []openai.ImageResponseDataInner{
			{
				URL:     "test-url1",
				B64JSON: "",
			},
			{
				URL:     "test-url2",
				B64JSON: "",
			},
			{
				URL:     "test-url3",
				B64JSON: "",
			},
		},
	}

	resBytes, _ = json.Marshal(responses)
	fmt.Fprintln(w, string(resBytes))
}

func TestImageVariation(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/images/variations", handleVariateImageEndpoint)

	origin, err := os.Create(filepath.Join(t.TempDir(), "image.png"))
	if err != nil {
		t.Fatalf("open origin file error: %v", err)
	}
	defer origin.Close()

	_, err = client.CreateVariImage(context.Background(), openai.ImageVariRequest{
		Image:          origin,
		N:              3,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
	})
	checks.NoError(t, err, "CreateImage error")
}

// handleVariateImageEndpoint Handles the images endpoint by the test server.
func handleVariateImageEndpoint(w http.ResponseWriter, r *http.Request) {
	var resBytes []byte

	// images only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	responses := openai.ImageResponse{
		Created: time.Now().Unix(),
		Data: []openai.ImageResponseDataInner{
			{
				URL:     "test-url1",
				B64JSON: "",
			},
			{
				URL:     "test-url2",
				B64JSON: "",
			},
			{
				URL:     "test-url3",
				B64JSON: "",
			},
		},
	}

	resBytes, _ = json.Marshal(responses)
	fmt.Fprintln(w, string(resBytes))
}
