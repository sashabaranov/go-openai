package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
		FilePath: "api.go",
		Purpose:  "fine-tune",
	}
	_, err = client.CreateFile(ctx, req)
	if err != nil {
		t.Fatalf("CreateFile error: %v", err)
	}
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
