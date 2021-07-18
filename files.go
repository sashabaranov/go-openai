package gogpt

import (
	"context"
	"fmt"
	"net/http"
)

// File struct represents an OpenAPI file
type File struct {
	Bytes     int    `json:"bytes"`
	CreatedAt int    `json:"created_at"`
	ID        string `json:"id"`
	FileName  string `json:"filename"`
	Object    string `json:"object"`
	Owner     string `json:"owner"`
	Purpose   string `json:"purpose"`
}

// FilesList is a list of files that belong to the user or organization
type FilesList struct {
	Files []File `json:"data"`
}

// ListFiles Lists the currently available files,
// and provides basic information about each file such as the file name and purpose.
func (c *Client) ListFiles(ctx context.Context) (files FilesList, err error) {
	req, err := http.NewRequest("GET", c.fullURL("/files"), nil)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &files)
	return
}

// GetFile Retrieves a file instance, providing basic information about the file
// such as the file name and purpose.
func (c *Client) GetFile(ctx context.Context, fileID string) (file File, err error) {
	urlSuffix := fmt.Sprintf("/files/%s", fileID)
	req, err := http.NewRequest("GET", c.fullURL(urlSuffix), nil)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	err = c.sendRequest(req, &file)
	return
}
