package openai //nolint:testpackage // testing private field

import (
	utils "github.com/sashabaranov/go-openai/internal"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"context"
	"encoding/json"
	"errors"
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

func TestDeleteFile(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/files/deadbeef", func(w http.ResponseWriter, r *http.Request) {

	})
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	err = client.DeleteFile(ctx, "deadbeef")
	checks.NoError(t, err, "DeleteFile error")
}

func TestListFile(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/files", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{}")
	})
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err = client.ListFiles(ctx)
	checks.NoError(t, err, "ListFiles error")
}

func TestGetFile(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/files/deadbeef", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{}")
	})
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err = client.GetFile(ctx, "deadbeef")
	checks.NoError(t, err, "GetFile error")
}

func TestGetFileContent(t *testing.T) {
	wantRespJsonl := `{"prompt": "foo", "completion": "foo"}
{"prompt": "bar", "completion": "bar"}
{"prompt": "baz", "completion": "baz"}
`
	server := test.NewTestServer()
	server.RegisterHandler("/v1/files/deadbeef/content", func(w http.ResponseWriter, r *http.Request) {
		// edits only accepts GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		fmt.Fprint(w, wantRespJsonl)
	})
	// create the test server
	var err error
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	content, err := client.GetFileContent(ctx, "deadbeef")
	checks.NoError(t, err, "GetFileContent error")
	defer content.Close()

	actual, _ := io.ReadAll(content)
	if string(actual) != wantRespJsonl {
		t.Errorf("Expected %s, got %s", wantRespJsonl, string(actual))
	}
}

func TestGetFileContentReturnError(t *testing.T) {
	wantMessage := "To help mitigate abuse, downloading of fine-tune training files is disabled for free accounts."
	wantType := "invalid_request_error"
	wantErrorResp := `{
  "error": {
    "message": "` + wantMessage + `",
    "type": "` + wantType + `",
    "param": null,
    "code": null
  }
}`
	server := test.NewTestServer()
	server.RegisterHandler("/v1/files/deadbeef/content", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, wantErrorResp)
	})
	// create the test server
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()

	_, err := client.GetFileContent(ctx, "deadbeef")
	if err == nil {
		t.Fatal("Did not return error")
	}

	apiErr := &APIError{}
	if !errors.As(err, &apiErr) {
		t.Fatalf("Did not return APIError: %+v\n", apiErr)
	}
	if apiErr.Message != wantMessage {
		t.Fatalf("Expected %s Message, got = %s\n", wantMessage, apiErr.Message)
		return
	}
	if apiErr.Type != wantType {
		t.Fatalf("Expected %s Type, got = %s\n", wantType, apiErr.Type)
		return
	}
}

func TestGetFileContentReturnTimeoutError(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/files/deadbeef/content", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Nanosecond)
	})
	// create the test server
	ts := server.OpenAITestServer()
	ts.Start()
	defer ts.Close()

	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client := NewClientWithConfig(config)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	defer cancel()

	_, err := client.GetFileContent(ctx, "deadbeef")
	if err == nil {
		t.Fatal("Did not return error")
	}
	if !os.IsTimeout(err) {
		t.Fatal("Did not return timeout error")
	}
}
