package openai //nolint:testpackage // testing private field

import (
	utils "github.com/sashabaranov/go-openai/internal"
	"github.com/sashabaranov/go-openai/internal/test"
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
	server := test.NewTestServer()
	server.RegisterHandler("/v1/images/generations", handleImageEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	req := ImageRequest{}
	req.Prompt = "Lorem ipsum"
	_, err = client.CreateImage(ctx, req)
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
	server := test.NewTestServer()
	server.RegisterHandler("/v1/images/edits", handleEditImageEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

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

	req := ImageEditRequest{
		Image:          origin,
		Mask:           mask,
		Prompt:         "There is a turtle in the pool",
		N:              3,
		Size:           CreateImageSize1024x1024,
		ResponseFormat: CreateImageResponseFormatURL,
	}
	_, err = client.CreateEditImage(ctx, req)
	checks.NoError(t, err, "CreateImage error")
}

func TestImageEditWithoutMask(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/images/edits", handleEditImageEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	origin, err := os.Create("image.png")
	if err != nil {
		t.Error("open origin file error")
		return
	}

	defer func() {
		origin.Close()
		os.Remove("image.png")
	}()

	req := ImageEditRequest{
		Image:          origin,
		Prompt:         "There is a turtle in the pool",
		N:              3,
		Size:           CreateImageSize1024x1024,
		ResponseFormat: CreateImageResponseFormatURL,
	}
	_, err = client.CreateEditImage(ctx, req)
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
	server := test.NewTestServer()
	server.RegisterHandler("/v1/images/variations", handleVariateImageEndpoint)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	origin, err := os.Create("image.png")
	if err != nil {
		t.Error("open origin file error")
		return
	}

	defer func() {
		origin.Close()
		os.Remove("image.png")
	}()

	req := ImageVariRequest{
		Image:          origin,
		N:              3,
		Size:           CreateImageSize1024x1024,
		ResponseFormat: CreateImageResponseFormatURL,
	}
	_, err = client.CreateVariImage(ctx, req)
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

type mockFormBuilder struct {
	mockCreateFormFile func(string, *os.File) error
	mockWriteField     func(string, string) error
	mockClose          func() error
}

func (fb *mockFormBuilder) CreateFormFile(fieldname string, file *os.File) error {
	return fb.mockCreateFormFile(fieldname, file)
}

func (fb *mockFormBuilder) WriteField(fieldname, value string) error {
	return fb.mockWriteField(fieldname, value)
}

func (fb *mockFormBuilder) Close() error {
	return fb.mockClose()
}

func (fb *mockFormBuilder) FormDataContentType() string {
	return ""
}

func TestImageFormBuilderFailures(t *testing.T) {
	config := DefaultConfig("")
	config.BaseURL = ""
	client := NewClientWithConfig(config)

	mockBuilder := &mockFormBuilder{}
	client.createFormBuilder = func(io.Writer) utils.FormBuilder {
		return mockBuilder
	}
	ctx := context.Background()

	req := ImageEditRequest{
		Mask: &os.File{},
	}

	mockFailedErr := fmt.Errorf("mock form builder fail")
	mockBuilder.mockCreateFormFile = func(string, *os.File) error {
		return mockFailedErr
	}
	_, err := client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	mockBuilder.mockCreateFormFile = func(name string, file *os.File) error {
		if name == "mask" {
			return mockFailedErr
		}
		return nil
	}
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	mockBuilder.mockCreateFormFile = func(name string, file *os.File) error {
		return nil
	}

	var failForField string
	mockBuilder.mockWriteField = func(fieldname, value string) error {
		if fieldname == failForField {
			return mockFailedErr
		}
		return nil
	}

	failForField = "prompt"
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	failForField = "n"
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	failForField = "size"
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	failForField = "response_format"
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")

	failForField = ""
	mockBuilder.mockClose = func() error {
		return mockFailedErr
	}
	_, err = client.CreateEditImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")
}

func TestVariImageFormBuilderFailures(t *testing.T) {
	config := DefaultConfig("")
	config.BaseURL = ""
	client := NewClientWithConfig(config)

	mockBuilder := &mockFormBuilder{}
	client.createFormBuilder = func(io.Writer) utils.FormBuilder {
		return mockBuilder
	}
	ctx := context.Background()

	req := ImageVariRequest{}

	mockFailedErr := fmt.Errorf("mock form builder fail")
	mockBuilder.mockCreateFormFile = func(string, *os.File) error {
		return mockFailedErr
	}
	_, err := client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateVariImage should return error if form builder fails")

	mockBuilder.mockCreateFormFile = func(name string, file *os.File) error {
		return nil
	}

	var failForField string
	mockBuilder.mockWriteField = func(fieldname, value string) error {
		if fieldname == failForField {
			return mockFailedErr
		}
		return nil
	}

	failForField = "n"
	_, err = client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateVariImage should return error if form builder fails")

	failForField = "size"
	_, err = client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateVariImage should return error if form builder fails")

	failForField = "response_format"
	_, err = client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateVariImage should return error if form builder fails")

	failForField = ""
	mockBuilder.mockClose = func() error {
		return mockFailedErr
	}
	_, err = client.CreateVariImage(ctx, req)
	checks.ErrorIs(t, err, mockFailedErr, "CreateImage should return error if form builder fails")
}
