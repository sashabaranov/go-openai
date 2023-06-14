package openai_test

import (
	"bytes"
	"strings"

	. "github.com/coggsfl/go-openai"
	"github.com/coggsfl/go-openai/internal/test/checks"

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

func TestAzureImages(t *testing.T) {
	client, server, teardown := setupAzureTestServer()
	defer teardown()
	server.RegisterHandler("/openai/images/generations:submit", handleImageEndpoint)
	server.RegisterHandler("/openai/operations/images/request-id", handleImageCallbackEndpoint)

	_, err := client.CreateImage(context.Background(), ImageRequest{
		Prompt:         "Lorem ipsum",
		ResponseFormat: CreateImageResponseFormatURL,
		N:              2,
	})
	checks.NoError(t, err, "Azure CreateImage error")
}

// handleImageEndpoint Handles the images endpoint by the test server.
func handleImageEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// imagess only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Azure Image Generation request - respond with callback Header only & HTTP accepted status.
	if strings.Contains(r.RequestURI, "/openai/images/generations:submit") {
		w.Header().Add("Operation-Location", "http://"+r.Host+"/openai/operations/images/request-id")
		w.WriteHeader(http.StatusAccepted)
		return
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

// handleImageCallbackEndpoint Handles the callback endpoint by the test server.
func handleImageCallbackEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error

	// image callback only accepts GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set the status to succeeded if this is a retry request.
	status := "running"
	if r.Header.Get("Retry") == "true" {
		status = "succeeded"
	}

	cbResponse := CallBackResponse{
		Created: time.Now().Unix(),
		Status:  status,
		Result: CBResult{
			Data: CBData{
				{URL: "http://example.com/image1"},
				{URL: "http://example.com/image2"},
			},
		},
	}
	cbResponseBytes := new(bytes.Buffer)
	err = json.NewEncoder(cbResponseBytes).Encode(cbResponse)
	if err != nil {
		http.Error(w, "could not write repsonse", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, cbResponseBytes.String())
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
