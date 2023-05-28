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
	"strconv"
	"testing"
	"time"
)

func TestFileUpload(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/files", handleCreateFile)
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	req := FileRequest{
		FileName: "test.go",
		FilePath: "client.go",
		Purpose:  "fine-tune",
	}
	_, err = client.CreateFile(ctx, req)
	checks.NoError(t, err, "CreateFile error")
}

// handleCreateFile Handles the images endpoint by the test server.
func handleCreateFile(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// edits only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	err = r.ParseMultipartForm(1024 * 1024 * 1024)
	if err != nil {
		http.Error(w, "file is more than 1GB", http.StatusInternalServerError)
		return
	}

	values := r.Form
	var purpose string
	for key, value := range values {
		if key == "purpose" {
			purpose = value[0]
		}
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return
	}
	defer file.Close()

	var fileReq = File{
		Bytes:     int(header.Size),
		ID:        strconv.Itoa(int(time.Now().Unix())),
		FileName:  header.Filename,
		Purpose:   purpose,
		CreatedAt: time.Now().Unix(),
		Object:    "test-objecct",
		Owner:     "test-owner",
	}

	resBytes, _ = json.Marshal(fileReq)
	fmt.Fprint(w, string(resBytes))
}

func TestFileUploadWithFailingFormBuilder(t *testing.T) {
	config := DefaultConfig("")
	config.BaseURL = ""
	client := NewClientWithConfig(config)
	mockBuilder := &mockFormBuilder{}
	client.createFormBuilder = func(io.Writer) utils.FormBuilder {
		return mockBuilder
	}

	ctx := context.Background()
	req := FileRequest{
		FileName: "test.go",
		FilePath: "client.go",
		Purpose:  "fine-tune",
	}

	mockError := fmt.Errorf("mockWriteField error")
	mockBuilder.mockWriteField = func(string, string) error {
		return mockError
	}
	_, err := client.CreateFile(ctx, req)
	checks.ErrorIs(t, err, mockError, "CreateFile should return error if form builder fails")

	mockError = fmt.Errorf("mockCreateFormFile error")
	mockBuilder.mockWriteField = func(string, string) error {
		return nil
	}
	mockBuilder.mockCreateFormFile = func(string, *os.File) error {
		return mockError
	}
	_, err = client.CreateFile(ctx, req)
	checks.ErrorIs(t, err, mockError, "CreateFile should return error if form builder fails")

	mockError = fmt.Errorf("mockClose error")
	mockBuilder.mockWriteField = func(string, string) error {
		return nil
	}
	mockBuilder.mockCreateFormFile = func(string, *os.File) error {
		return nil
	}
	mockBuilder.mockClose = func() error {
		return mockError
	}
	_, err = client.CreateFile(ctx, req)
	checks.ErrorIs(t, err, mockError, "CreateFile should return error if form builder fails")
}

func TestFileUploadWithNonExistentPath(t *testing.T) {
	config := DefaultConfig("")
	config.BaseURL = ""
	client := NewClientWithConfig(config)

	ctx := context.Background()
	req := FileRequest{
		FilePath: "some non existent file path/F616FD18-589E-44A8-BF0C-891EAE69C455",
	}

	_, err := client.CreateFile(ctx, req)
	checks.ErrorIs(t, err, os.ErrNotExist, "CreateFile should return error if file does not exist")
}
