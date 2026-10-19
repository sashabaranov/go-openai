package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/url"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

type vectorStoreErrorTransport func(*http.Request) (*http.Response, error)

func (f vectorStoreErrorTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestVectorStoreInvalidRequestURL(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name string
		call func(*openai.Client) error
	}{
		{"create", func(c *openai.Client) error {
			_, err := c.CreateVectorStore(ctx, openai.VectorStoreRequest{})
			return err
		}},
		{"retrieve", func(c *openai.Client) error {
			_, err := c.RetrieveVectorStore(ctx, "vs-test")
			return err
		}},
		{"modify", func(c *openai.Client) error {
			_, err := c.ModifyVectorStore(ctx, "vs-test", openai.VectorStoreRequest{})
			return err
		}},
		{"delete", func(c *openai.Client) error {
			_, err := c.DeleteVectorStore(ctx, "vs-test")
			return err
		}},
		{"list", func(c *openai.Client) error {
			_, err := c.ListVectorStores(ctx, openai.Pagination{})
			return err
		}},
		{"create file", func(c *openai.Client) error {
			_, err := c.CreateVectorStoreFile(ctx, "vs-test", openai.VectorStoreFileRequest{})
			return err
		}},
		{"retrieve file", func(c *openai.Client) error {
			_, err := c.RetrieveVectorStoreFile(ctx, "vs-test", "file-test")
			return err
		}},
		{"delete file", func(c *openai.Client) error {
			return c.DeleteVectorStoreFile(ctx, "vs-test", "file-test")
		}},
		{"list files", func(c *openai.Client) error {
			_, err := c.ListVectorStoreFiles(ctx, "vs-test", openai.Pagination{})
			return err
		}},
		{"create batch", func(c *openai.Client) error {
			_, err := c.CreateVectorStoreFileBatch(ctx, "vs-test", openai.VectorStoreFileBatchRequest{})
			return err
		}},
		{"retrieve batch", func(c *openai.Client) error {
			_, err := c.RetrieveVectorStoreFileBatch(ctx, "vs-test", "batch-test")
			return err
		}},
		{"cancel batch", func(c *openai.Client) error {
			_, err := c.CancelVectorStoreFileBatch(ctx, "vs-test", "batch-test")
			return err
		}},
		{"list batch files", func(c *openai.Client) error {
			_, err := c.ListVectorStoreFilesInBatch(ctx, "vs-test", "batch-test", openai.Pagination{})
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("request construction panicked: %v", p)
				}
			}()
			calls := 0
			config := openai.DefaultConfig("test-key")
			config.BaseURL = "https://example.invalid/%zz"
			config.HTTPClient = &http.Client{Transport: vectorStoreErrorTransport(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("unexpected HTTP request")
			})}
			err := tc.call(openai.NewClientWithConfig(config))
			var urlErr *url.Error
			if !errors.As(err, &urlErr) {
				t.Errorf("got %v, want the URL construction error", err)
			}
			if calls != 0 {
				t.Errorf("HTTP calls = %d, want 0", calls)
			}
		})
	}
}

func TestVectorStoreInvalidMetadata(t *testing.T) {
	ctx := context.Background()
	request := openai.VectorStoreRequest{Metadata: map[string]any{"value": math.NaN()}}
	for _, tc := range []struct {
		name string
		call func(*openai.Client) error
	}{
		{"create", func(c *openai.Client) error {
			_, err := c.CreateVectorStore(ctx, request)
			return err
		}},
		{"modify", func(c *openai.Client) error {
			_, err := c.ModifyVectorStore(ctx, "vs-test", request)
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("metadata encoding panicked: %v", p)
				}
			}()
			calls := 0
			config := openai.DefaultConfig("test-key")
			config.HTTPClient = &http.Client{Transport: vectorStoreErrorTransport(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("unexpected HTTP request")
			})}
			err := tc.call(openai.NewClientWithConfig(config))
			var encodingErr *json.UnsupportedValueError
			if !errors.As(err, &encodingErr) {
				t.Errorf("got %v, want the metadata encoding error", err)
			}
			if calls != 0 {
				t.Errorf("HTTP calls = %d, want 0", calls)
			}
		})
	}
}
