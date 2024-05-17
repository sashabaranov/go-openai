package openai_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestCreateBatch(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files", handleCreateFile)
	server.RegisterHandler("/v1/batches", handleBatchEndpoint)
	req := openai.CreateBatchRequest{
		Endpoint: openai.BatchEndpointChatCompletions,
	}
	req.AddChatCompletion("req-1", openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	})
	_, err := client.CreateBatch(context.Background(), req)
	checks.NoError(t, err, "CreateBatch error")
}

func TestRetrieveBatch(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/batches/file-id-1", handleRetrieveBatchEndpoint)
	_, err := client.RetrieveBatch(context.Background(), "file-id-1")
	checks.NoError(t, err, "RetrieveBatch error")
}

func TestCancelBatch(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/batches/file-id-1/cancel", handleCancelBatchEndpoint)
	_, err := client.CancelBatch(context.Background(), "file-id-1")
	checks.NoError(t, err, "RetrieveBatch error")
}

func TestListBatch(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/batches", handleBatchEndpoint)
	_, err := client.ListBatch(context.Background(), nil, nil)
	checks.NoError(t, err, "RetrieveBatch error")
}

func handleBatchEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_, _ = fmt.Fprintln(w, `{
			  "id": "batch_abc123",
			  "object": "batch",
			  "endpoint": "/v1/completions",
			  "errors": null,
			  "input_file_id": "file-abc123",
			  "completion_window": "24h",
			  "status": "completed",
			  "output_file_id": "file-cvaTdG",
			  "error_file_id": "file-HOWS94",
			  "created_at": 1711471533,
			  "in_progress_at": 1711471538,
			  "expires_at": 1711557933,
			  "finalizing_at": 1711493133,
			  "completed_at": 1711493163,
			  "failed_at": null,
			  "expired_at": null,
			  "cancelling_at": null,
			  "cancelled_at": null,
			  "request_counts": {
				"total": 100,
				"completed": 95,
				"failed": 5
			  },
			  "metadata": {
				"customer_id": "user_123456789",
				"batch_description": "Nightly eval job"
			  }
			}`)
	} else if r.Method == http.MethodGet {
		_, _ = fmt.Fprintln(w, `{
			  "object": "list",
			  "data": [
				{
				  "id": "batch_abc123",
				  "object": "batch",
				  "endpoint": "/v1/chat/completions",
				  "errors": null,
				  "input_file_id": "file-abc123",
				  "completion_window": "24h",
				  "status": "completed",
				  "output_file_id": "file-cvaTdG",
				  "error_file_id": "file-HOWS94",
				  "created_at": 1711471533,
				  "in_progress_at": 1711471538,
				  "expires_at": 1711557933,
				  "finalizing_at": 1711493133,
				  "completed_at": 1711493163,
				  "failed_at": null,
				  "expired_at": null,
				  "cancelling_at": null,
				  "cancelled_at": null,
				  "request_counts": {
					"total": 100,
					"completed": 95,
					"failed": 5
				  },
				  "metadata": {
					"customer_id": "user_123456789",
					"batch_description": "Nightly job"
				  }
				}
			  ],
			  "first_id": "batch_abc123",
			  "last_id": "batch_abc456",
			  "has_more": true
			}`)
	}
}

func handleRetrieveBatchEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		_, _ = fmt.Fprintln(w, `{
		  "id": "batch_abc123",
		  "object": "batch",
		  "endpoint": "/v1/completions",
		  "errors": null,
		  "input_file_id": "file-abc123",
		  "completion_window": "24h",
		  "status": "completed",
		  "output_file_id": "file-cvaTdG",
		  "error_file_id": "file-HOWS94",
		  "created_at": 1711471533,
		  "in_progress_at": 1711471538,
		  "expires_at": 1711557933,
		  "finalizing_at": 1711493133,
		  "completed_at": 1711493163,
		  "failed_at": null,
		  "expired_at": null,
		  "cancelling_at": null,
		  "cancelled_at": null,
		  "request_counts": {
			"total": 100,
			"completed": 95,
			"failed": 5
		  },
		  "metadata": {
			"customer_id": "user_123456789",
			"batch_description": "Nightly eval job"
		  }
		}`)
	}
}

func handleCancelBatchEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_, _ = fmt.Fprintln(w, `{
		  "id": "batch_abc123",
		  "object": "batch",
		  "endpoint": "/v1/chat/completions",
		  "errors": null,
		  "input_file_id": "file-abc123",
		  "completion_window": "24h",
		  "status": "cancelling",
		  "output_file_id": null,
		  "error_file_id": null,
		  "created_at": 1711471533,
		  "in_progress_at": 1711471538,
		  "expires_at": 1711557933,
		  "finalizing_at": null,
		  "completed_at": null,
		  "failed_at": null,
		  "expired_at": null,
		  "cancelling_at": 1711475133,
		  "cancelled_at": null,
		  "request_counts": {
			"total": 100,
			"completed": 23,
			"failed": 1
		  },
		  "metadata": {
			"customer_id": "user_123456789",
			"batch_description": "Nightly eval job"
		  }
		}`)
	}
}
