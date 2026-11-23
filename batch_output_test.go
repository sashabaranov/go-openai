package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestBatchRequestOutputChatCompletion(t *testing.T) {
	body := decodeBatchOutputBody[openai.ChatCompletionResponse](t, `{
		"id":"chatcmpl-123","object":"chat.completion","created":1711652795,"model":"gpt-4o-mini",
		"choices":[{"index":0,"message":{"role":"assistant","content":"Hello."},"finish_reason":"stop"}],
		"usage":{"prompt_tokens":22,"completion_tokens":2,"total_tokens":24},"system_fingerprint":"fp_123"
	}`)
	if body.ID != "chatcmpl-123" || len(body.Choices) != 1 {
		t.Fatalf("unexpected chat completion body: %#v", body)
	}
	if body.Choices[0].Message.Content != "Hello." || body.Usage.TotalTokens != 24 {
		t.Errorf("unexpected chat completion content or usage: %#v", body)
	}
}

func TestBatchRequestOutputCompletion(t *testing.T) {
	body := decodeBatchOutputBody[openai.CompletionResponse](t, `{
		"id":"cmpl-123","object":"text_completion","created":1711652795,"model":"davinci-002",
		"choices":[{"index":0,"text":"Hello.","finish_reason":"stop"}],
		"usage":{"prompt_tokens":22,"completion_tokens":2,"total_tokens":24}
	}`)
	if body.ID != "cmpl-123" || len(body.Choices) != 1 || body.Usage == nil {
		t.Fatalf("unexpected completion body: %#v", body)
	}
	if body.Choices[0].Text != "Hello." || body.Usage.TotalTokens != 24 {
		t.Errorf("unexpected completion text or usage: %#v", body)
	}
}

func TestBatchRequestOutputEmbedding(t *testing.T) {
	body := decodeBatchOutputBody[openai.EmbeddingResponse](t, `{
		"object":"list","data":[{"object":"embedding","index":0,"embedding":[0.25,-0.5]}],
		"model":"text-embedding-3-small","usage":{"prompt_tokens":2,"total_tokens":2}
	}`)
	if body.Model != openai.SmallEmbedding3 || len(body.Data) != 1 {
		t.Fatalf("unexpected embedding body: %#v", body)
	}
	if !reflect.DeepEqual(body.Data[0].Embedding, []float32{0.25, -0.5}) || body.Usage.TotalTokens != 2 {
		t.Errorf("unexpected embedding values or usage: %#v", body)
	}
}

func TestBatchRequestOutputResponse(t *testing.T) {
	body := decodeBatchOutputBody[openai.CreateResponseResponse](t, `{
		"id":"resp-123","object":"response","created_at":1711652795,"status":"completed","model":"gpt-4o-mini",
		"output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello."}]}],
		"usage":{"input_tokens":22,"output_tokens":2,"total_tokens":24}
	}`)
	if body.ID != "resp-123" || body.GetOutputText() != "Hello." || body.Usage == nil {
		t.Fatalf("unexpected Responses API body: %#v", body)
	}
	if body.Status != openai.ResponseStatusCompleted || body.Usage.TotalTokens != 24 {
		t.Errorf("unexpected Responses API status or usage: %#v", body)
	}
}

func decodeBatchOutputBody[T any](t *testing.T, body string) T {
	t.Helper()
	data := `{"id":"batch_req_123","custom_id":"request-2",` +
		`"response":{"status_code":200,"request_id":"req_123","body":` + body + `},"error":null}`
	var output openai.BatchRequestOutput[T]
	checks.NoErrorF(t, json.Unmarshal([]byte(data), &output), "decode batch output")
	if output.ID != "batch_req_123" || output.CustomID != "request-2" || output.Error != nil || output.Response == nil {
		t.Fatalf("unexpected batch output: %#v", output)
	}
	if output.Response.StatusCode != http.StatusOK || output.Response.RequestID != "req_123" {
		t.Errorf("unexpected batch response: %#v", output.Response)
	}
	encoded, err := json.Marshal(output)
	checks.NoErrorF(t, err, "encode batch output")
	var roundtrip openai.BatchRequestOutput[T]
	checks.NoErrorF(t, json.Unmarshal(encoded, &roundtrip), "decode batch output again")
	if !reflect.DeepEqual(output, roundtrip) {
		t.Errorf("batch output changed after roundtrip: %s", encoded)
	}
	var fields map[string]json.RawMessage
	checks.NoErrorF(t, json.Unmarshal(encoded, &fields), "decode batch output fields")
	if string(fields["error"]) != "null" {
		t.Errorf("expected null error, got %s", fields["error"])
	}
	return output.Response.Body
}

func TestBatchRequestOutputError(t *testing.T) {
	data := `{"id":"batch_req_123","custom_id":"request-3","response":null,` +
		`"error":{"code":"batch_expired","message":"The completion window expired."}}`
	var output openai.BatchRequestOutput[openai.ChatCompletionResponse]
	checks.NoErrorF(t, json.Unmarshal([]byte(data), &output), "decode batch error")
	if output.ID != "batch_req_123" || output.CustomID != "request-3" || output.Response != nil || output.Error == nil {
		t.Fatalf("unexpected batch error output: %#v", output)
	}
	if output.Error.Code != "batch_expired" || output.Error.Message != "The completion window expired." {
		t.Errorf("unexpected batch error: %#v", output.Error)
	}
	encoded, err := json.Marshal(output)
	checks.NoErrorF(t, err, "encode batch error")
	if string(encoded) != data {
		t.Errorf("batch error changed after roundtrip: %s", encoded)
	}
}

func TestBatchRequestOutputAPIError(t *testing.T) {
	data := `{"id":"batch_req_456","custom_id":"request-1",` +
		`"response":{"status_code":400,"request_id":"req_456","body":{"error":{` +
		`"code":"invalid_value","message":"Invalid model.","param":"model","type":"invalid_request_error"}}},"error":null}`
	var output openai.BatchRequestOutput[openai.ErrorResponse]
	checks.NoErrorF(t, json.Unmarshal([]byte(data), &output), "decode API error output")
	if output.Error != nil || output.Response == nil || output.Response.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected API error output: %#v", output)
	}
	if output.ID != "batch_req_456" || output.CustomID != "request-1" || output.Response.RequestID != "req_456" {
		t.Errorf("unexpected API error identifiers: %#v", output)
	}
	apiError := output.Response.Body.Error
	if apiError == nil || apiError.Param == nil {
		t.Fatalf("missing API error or parameter: %#v", apiError)
	}
	if apiError.Code != "invalid_value" || apiError.Message != "Invalid model." ||
		*apiError.Param != "model" || apiError.Type != "invalid_request_error" {
		t.Errorf("unexpected API error body: %#v", apiError)
	}
}

func TestBatchRequestOutputFileContent(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	lines := []string{
		`{"id":"batch_req_2","custom_id":"request-2","response":{"status_code":200,"request_id":"req_2",` +
			`"body":{"id":"chatcmpl-2","choices":[{"message":{"role":"assistant","content":"Hello."}}]}},"error":null}`,
		`{"id":"batch_req_1","custom_id":"request-1","response":{"status_code":400,"request_id":"req_1",` +
			`"body":{"error":{"code":"invalid_value","message":"Invalid model.","type":"invalid_request_error"}}},"error":null}`,
		`{"id":"batch_req_3","custom_id":"request-3","response":null,` +
			`"error":{"code":"batch_expired","message":"The completion window expired."}}`,
	}
	server.RegisterHandler("/v1/files/file-output/content", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/jsonl")
		for _, line := range lines {
			_, err := fmt.Fprintln(w, line)
			checks.NoError(t, err, "write batch output line")
		}
	})
	content, err := client.GetFileContent(context.Background(), "file-output")
	checks.NoErrorF(t, err, "get batch output file")
	defer content.Close()

	decoder := json.NewDecoder(content)
	for _, line := range lines {
		var output openai.BatchRequestOutput[json.RawMessage]
		checks.NoErrorF(t, decoder.Decode(&output), "decode JSONL output")
		encoded, encodeErr := json.Marshal(output)
		checks.NoErrorF(t, encodeErr, "encode JSONL output")
		if string(encoded) != line {
			t.Errorf("JSONL output changed after roundtrip: %s", encoded)
		}
		switch output.CustomID {
		case "request-2":
			var body openai.ChatCompletionResponse
			checks.NoErrorF(t, json.Unmarshal(output.Response.Body, &body), "decode successful body")
			if body.ID != "chatcmpl-2" || len(body.Choices) != 1 || body.Choices[0].Message.Content != "Hello." {
				t.Errorf("unexpected chat completion: %#v", body)
			}
		case "request-1":
			var body openai.ErrorResponse
			checks.NoErrorF(t, json.Unmarshal(output.Response.Body, &body), "decode failed body")
			if body.Error == nil || body.Error.Code != "invalid_value" {
				t.Errorf("unexpected API error: %#v", body)
			}
		case "request-3":
			if output.Response != nil || output.Error == nil || output.Error.Code != "batch_expired" {
				t.Errorf("unexpected batch error: %#v", output)
			}
		default:
			t.Errorf("unexpected custom ID: %s", output.CustomID)
		}
	}
	var extra openai.BatchRequestOutput[json.RawMessage]
	checks.ErrorIs(t, decoder.Decode(&extra), io.EOF, "expected end of JSONL file")
}
