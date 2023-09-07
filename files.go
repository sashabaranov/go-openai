package openai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

type FileRequest struct {
	FileName string `json:"file"`
	FilePath string `json:"-"`
	Purpose  string `json:"purpose"`
}

// File struct represents an OpenAPI file.
type File struct {
	Bytes         int    `json:"bytes"`
	CreatedAt     int64  `json:"created_at"`
	ID            string `json:"id"`
	FileName      string `json:"filename"`
	Object        string `json:"object"`
	Status        string `json:"status"`
	Purpose       string `json:"purpose"`
	StatusDetails string `json:"status_details"`
}

// FilesList is a list of files that belong to the user or organization.
type FilesList struct {
	Files []File `json:"data"`
}

// CreateFile uploads a jsonl file to GPT3
// FilePath must be a local file path.
func (c *Client) CreateFile(ctx context.Context, request FileRequest) (file File, err error) {
	var b bytes.Buffer
	builder := c.createFormBuilder(&b)

	err = builder.WriteField("purpose", request.Purpose)
	if err != nil {
		return
	}

	fileData, err := os.Open(request.FilePath)
	if err != nil {
		return
	}

	err = builder.CreateFormFile("file", fileData)
	if err != nil {
		return
	}

	err = builder.Close()
	if err != nil {
		return
	}

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/files"),
		withBody(&b), withContentType(builder.FormDataContentType()))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &file)
	return
}

// DeleteFile deletes an existing file.
func (c *Client) DeleteFile(ctx context.Context, fileID string) (err error) {
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL("/files/"+fileID))
	if err != nil {
		return
	}

	err = c.sendRequest(req, nil)
	return
}

// ListFiles Lists the currently available files,
// and provides basic information about each file such as the file name and purpose.
func (c *Client) ListFiles(ctx context.Context) (files FilesList, err error) {
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL("/files"))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &files)
	return
}

// GetFile Retrieves a file instance, providing basic information about the file
// such as the file name and purpose.
func (c *Client) GetFile(ctx context.Context, fileID string) (file File, err error) {
	urlSuffix := fmt.Sprintf("/files/%s", fileID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &file)
	return
}

func (c *Client) GetFileContent(ctx context.Context, fileID string) (content io.ReadCloser, err error) {
	urlSuffix := fmt.Sprintf("/files/%s/content", fileID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	content, err = c.sendRequestRaw(req)
	return
}
