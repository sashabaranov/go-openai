package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestImages(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/images/generations", handleImageEndpoint)
	_, err := client.CreateImage(context.Background(), ImageRequest{
		Prompt: "Lorem ipsum",
	})
	checks.NoError(t, err, "CreateImage error")
}

// handleImageEndpoint Handles the images endpoint by the test server.
func handleImageEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// imagess only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	var imageReq ImageRequest
	if imageReq, err = getImageBody(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}
	res := ImageResponse{
		Created: time.Now().Unix(),
	}
	for i := 0; i < imageReq.N; i++ {
		imageData := ImageResponseDataInner{}
		switch imageReq.ResponseFormat {
		case CreateImageResponseFormatURL, "":
			imageData.URL = "https://example.com/image.png"
		case CreateImageResponseFormatB64JSON:
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
func getImageBody(r *http.Request) (ImageRequest, error) {
	image := ImageRequest{}
	// read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return ImageRequest{}, err
	}
	err = json.Unmarshal(reqBody, &image)
	if err != nil {
		return ImageRequest{}, err
	}
	return image, nil
}

func TestImageEdit(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/images/edits", handleEditImageEndpoint)

	origin, err := os.Create("image.png")
	if err != nil {
		t.Error("open origin file error")
		return
	}

	mask, err := os.Create("mask.png")
	if err != nil {
		t.Error("open mask file error")
		return
	}

	defer func() {
		mask.Close()
		origin.Close()
		os.Remove("mask.png")
		os.Remove("image.png")
	}()

	_, err = client.CreateEditImage(context.Background(), ImageEditRequest{
		Image:          origin,
		Mask:           mask,
		Prompt:         "There is a turtle in the pool",
		N:              3,
		Size:           CreateImageSize1024x1024,
		ResponseFormat: CreateImageResponseFormatURL,
	})
	checks.NoError(t, err, "CreateImage error")
}

func TestImageEditWithoutMask(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/images/edits", handleEditImageEndpoint)

	origin, err := os.Create("image.png")
	if err != nil {
		t.Error("open origin file error")
		return
	}

	defer func() {
		origin.Close()
		os.Remove("image.png")
	}()

	_, err = client.CreateEditImage(context.Background(), ImageEditRequest{
		Image:          origin,
		Prompt:         "There is a turtle in the pool",
		N:              3,
		Size:           CreateImageSize1024x1024,
		ResponseFormat: CreateImageResponseFormatURL,
	})
	checks.NoError(t, err, "CreateImage error")
}

// handleEditImageEndpoint Handles the images endpoint by the test server.
func handleEditImageEndpoint(w http.ResponseWriter, r *http.Request) {
	var resBytes []byte

	// imagess only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	responses := ImageResponse{
		Created: time.Now().Unix(),
		Data: []ImageResponseDataInner{
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

	origin, err := os.Create("image.png")
	if err != nil {
		t.Error("open origin file error")
		return
	}

	defer func() {
		origin.Close()
		os.Remove("image.png")
	}()

	_, err = client.CreateVariImage(context.Background(), ImageVariRequest{
		Image:          origin,
		N:              3,
		Size:           CreateImageSize1024x1024,
		ResponseFormat: CreateImageResponseFormatURL,
	})
	checks.NoError(t, err, "CreateImage error")
}

// handleVariateImageEndpoint Handles the images endpoint by the test server.
func handleVariateImageEndpoint(w http.ResponseWriter, r *http.Request) {
	var resBytes []byte

	// imagess only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	responses := ImageResponse{
		Created: time.Now().Unix(),
		Data: []ImageResponseDataInner{
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
