package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestCreateResponseStream(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/responses", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("expected text/event-stream Accept header, got %q", r.Header.Get("Accept"))
		}
		var request openai.CreateResponseRequest
		checks.NoError(t, json.NewDecoder(r.Body).Decode(&request), "decode streaming request")
		if !request.Stream {
			t.Errorf("expected stream=true")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set(xCustomHeader, xCustomHeaderValue)
		data := "event: response.created\n" +
			"data: {\"type\":\"response.created\",\"sequence_number\":0," +
			"\"response\":{\"id\":\"resp_123\",\"object\":\"response\",\"status\":\"in_progress\"," +
			"\"model\":\"gpt-4o\",\"output\":[]}}\n\n" +
			"event: response.output_text.delta\n" +
			"data: {\"type\":\"response.output_text.delta\",\"sequence_number\":1," +
			"\"item_id\":\"msg_123\",\"output_index\":0,\"content_index\":0,\"delta\":\"Hello\"}\n\n" +
			"event: response.completed\n" +
			"data: {\"type\":\"response.completed\",\"sequence_number\":2," +
			"\"response\":{\"id\":\"resp_123\",\"object\":\"response\",\"status\":\"completed\"," +
			"\"model\":\"gpt-4o\",\"output\":[],\"output_text\":\"Hello\"}}\n\n"
		_, err := w.Write([]byte(data))
		checks.NoError(t, err, "write stream")
	})

	stream, err := client.CreateResponseStream(context.Background(), openai.CreateResponseRequest{
		Model: openai.GPT4o,
		Input: "Hello",
	})
	checks.NoError(t, err, "CreateResponseStream returned error")
	defer stream.Close()

	if stream.Header().Get(xCustomHeader) != xCustomHeaderValue {
		t.Errorf("expected stream response header to be retained")
	}

	created, err := stream.Recv()
	checks.NoError(t, err, "receive created event")
	if created.Type != openai.ResponseStreamEventCreated || created.Response == nil {
		t.Fatalf("unexpected created event: %#v", created)
	}
	if created.Response.Status != openai.ResponseStatusInProgress {
		t.Errorf("unexpected created response status: %q", created.Response.Status)
	}

	delta, err := stream.Recv()
	checks.NoError(t, err, "receive delta event")
	if delta.Type != openai.ResponseStreamEventOutputTextDelta || delta.Delta != "Hello" {
		t.Errorf("unexpected delta event: %#v", delta)
	}
	if len(delta.Raw) == 0 {
		t.Errorf("expected raw event payload to be retained")
	}

	completed, err := stream.Recv()
	checks.NoError(t, err, "receive completed event")
	if completed.Type != openai.ResponseStreamEventCompleted || completed.Response == nil {
		t.Fatalf("unexpected completed event: %#v", completed)
	}
	if completed.Response.GetOutputText() != "Hello" {
		t.Errorf("unexpected completed output: %q", completed.Response.GetOutputText())
	}

	_, err = stream.Recv()
	checks.ErrorIs(t, err, io.EOF, "expected stream EOF")
	_, err = stream.Recv()
	checks.ErrorIs(t, err, io.EOF, "expected repeated stream EOF")
}

func TestCreateResponseStreamHTTPError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/responses", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error":{"message":"bad request","type":"invalid_request_error"}}`))
		checks.NoError(t, err, "write error response")
	})

	stream, err := client.CreateResponseStream(context.Background(), openai.CreateResponseRequest{
		Model: openai.GPT4o,
		Input: "Hello",
	})
	if stream != nil {
		t.Errorf("expected nil stream on HTTP error")
	}
	var apiError *openai.APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
}

func TestCreateResponseStreamMalformedEvent(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/responses", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, err := w.Write([]byte("data: {not-json}\n\n"))
		checks.NoError(t, err, "write malformed stream")
	})

	stream, err := client.CreateResponseStream(context.Background(), openai.CreateResponseRequest{
		Model: openai.GPT4o,
		Input: "Hello",
	})
	checks.NoError(t, err, "CreateResponseStream returned error")
	defer stream.Close()

	_, err = stream.Recv()
	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Fatalf("expected JSON syntax error, got %T: %v", err, err)
	}
}
