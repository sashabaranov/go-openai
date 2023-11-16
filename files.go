package openai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type Purpose string

const (
	PurposeAssistants Purpose = "assistants"
	PurposeFineTune   Purpose = "fine-tune"
)

type FileRequest struct {
	Purpose  Purpose   `json:"purpose"`
	FilePath string    `json:"-"`
	FileName string    `json:"file"` // must be set if FilePath is not set
	File     io.Reader `json:"-"`    // also
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

	httpHeader
}

// FilesList is a list of files that belong to the user or organization.
type FilesList struct {
	Files []File `json:"data"`

	httpHeader
}

// CreateFile uploads a file that can be used across various endpoints.
//
// If the request does not set io.Reader, the client will try to open and read
// the file from local filesystem using FilePath. The size of all the files
// uploaded by one organization can be up to 100 GB. The size of individual
// files can be a maximum of 512 MB.
//
// The Fine-tuning API only supports .jsonl files.

func (c *Client) CreateFile(ctx context.Context, request FileRequest) (file File, err error) {
	fname := request.FileName
	if fname == "" {
		fname = path.Base(request.FilePath)
	}
	switch request.Purpose {
	case "":
		err = fmt.Errorf("openai: file with no purpose")
		return
	case PurposeFineTune:
		if strings.HasSuffix(fname, ".jsonl") {
			break
		}
		err = fmt.Errorf("openai: fine-tuning only supports .jsonl files")
		return
	}
	var b bytes.Buffer
	builder := c.createFormBuilder(&b)
	err = builder.WriteField("purpose", string(request.Purpose))
	if err != nil {
		return
	}
	switch r, fpath := request.File, request.FilePath; {
	case r != nil:
		err = builder.CreateFormFileReader("file", r, fname)
	case fpath != "":
		f, ret := os.Open(fpath)
		if ret != nil {
			return file, ret
		}
		defer f.Close()
		err = builder.CreateFormFileReader("file", f, fname)
	default:
		err = fmt.Errorf("openai: no reader or file path")
	}
	if err != nil {
		return
	}
	if err = builder.Close(); err != nil {
		return
	}
	req, ret := c.newRequest(ctx, http.MethodPost, c.fullURL("/files"),
		withBody(&b),
		withContentType(builder.FormDataContentType()))
	if ret != nil {
		return file, ret
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
