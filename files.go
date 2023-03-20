package openai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

type FileRequest struct {
	FileName string `json:"file"`
	FilePath string `json:"-"`
	Purpose  string `json:"purpose"`
}

// File struct represents an OpenAPI file.
type File struct {
	Bytes     int    `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	ID        string `json:"id"`
	FileName  string `json:"filename"`
	Object    string `json:"object"`
	Owner     string `json:"owner"`
	Purpose   string `json:"purpose"`
}

// FilesList is a list of files that belong to the user or organization.
type FilesList struct {
	Files []File `json:"data"`
}

// isUrl is a helper function that determines whether the given FilePath
// is a remote URL or a local file path.
func isURL(path string) bool {
	_, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}

	u, err := url.Parse(path)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// CreateFile uploads a jsonl file to GPT3
// FilePath can be either a local file path or a URL.
func (c *Client) CreateFile(ctx context.Context, request FileRequest) (file File, err error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	var fw io.Writer

	err = w.WriteField("purpose", request.Purpose)
	if err != nil {
		return
	}

	fw, err = w.CreateFormFile("file", request.FileName)
	if err != nil {
		return
	}

	var fileData io.ReadCloser
	if isURL(request.FilePath) {
		var remoteFile *http.Response
		remoteFile, err = http.Get(request.FilePath)
		if err != nil {
			return
		}

		defer remoteFile.Body.Close()

		// Check server response
		if remoteFile.StatusCode != http.StatusOK {
			err = fmt.Errorf("error, status code: %d, message: failed to fetch file", remoteFile.StatusCode)
			return
		}

		fileData = remoteFile.Body
	} else {
		fileData, err = os.Open(request.FilePath)
		if err != nil {
			return
		}
	}

	_, err = io.Copy(fw, fileData)
	if err != nil {
		return
	}

	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL("/files"), &b)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	err = c.sendRequest(req, &file)

	return
}

// DeleteFile deletes an existing file.
func (c *Client) DeleteFile(ctx context.Context, fileID string) (err error) {
	req, err := c.requestBuilder.build(ctx, http.MethodDelete, c.fullURL("/files/"+fileID), nil)
	if err != nil {
		return
	}

	err = c.sendRequest(req, nil)
	return
}

// ListFiles Lists the currently available files,
// and provides basic information about each file such as the file name and purpose.
func (c *Client) ListFiles(ctx context.Context) (files FilesList, err error) {
	req, err := c.requestBuilder.build(ctx, http.MethodGet, c.fullURL("/files"), nil)
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
	req, err := c.requestBuilder.build(ctx, http.MethodGet, c.fullURL(urlSuffix), nil)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &file)
	return
}
