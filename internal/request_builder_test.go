package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
)

var errTestMarshallerFailed = errors.New("test marshaller failed")

type failingMarshaller struct{}

func (*failingMarshaller) Marshal(_ any) ([]byte, error) {
	return []byte{}, errTestMarshallerFailed
}

func TestRequestBuilderReturnsMarshallerErrors(t *testing.T) {
	builder := HTTPRequestBuilder{
		marshaller: &failingMarshaller{},
	}

	_, err := builder.Build(context.Background(), "", "", struct{}{}, nil)
	if !errors.Is(err, errTestMarshallerFailed) {
		t.Fatalf("Did not return error when marshaller failed: %v", err)
	}
}

func TestRequestBuilderReturnsRequest(t *testing.T) {
	b := NewRequestBuilder()
	var (
		ctx         = context.Background()
		method      = http.MethodPost
		url         = "/foo"
		request     = map[string]string{"foo": "bar"}
		reqBytes, _ = b.marshaller.Marshal(request)
		want, _     = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBytes))
	)
	got, _ := b.Build(ctx, method, url, request, nil)
	if !reflect.DeepEqual(got.Body, want.Body) ||
		!reflect.DeepEqual(got.URL, want.URL) ||
		!reflect.DeepEqual(got.Method, want.Method) {
		t.Errorf("Build() got = %v, want %v", got, want)
	}
}

func TestRequestBuilderReturnsRequestWhenRequestOfArgsIsNil(t *testing.T) {
	var (
		ctx     = context.Background()
		method  = http.MethodGet
		url     = "/foo"
		want, _ = http.NewRequestWithContext(ctx, method, url, nil)
	)
	b := NewRequestBuilder()
	got, _ := b.Build(ctx, method, url, nil, nil)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build() got = %v, want %v", got, want)
	}
}

func TestRequestBuilderWithReaderBodyAndHeader(t *testing.T) {
	b := NewRequestBuilder()
	ctx := context.Background()
	method := http.MethodPost
	url := "/reader"
	bodyContent := "hello"
	body := bytes.NewBufferString(bodyContent)
	header := http.Header{"X-Test": []string{"val"}}

	req, err := b.Build(ctx, method, url, body, header)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	gotBody, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("cannot read body: %v", err)
	}
	if string(gotBody) != bodyContent {
		t.Fatalf("expected body %q, got %q", bodyContent, string(gotBody))
	}
	if req.Header.Get("X-Test") != "val" {
		t.Fatalf("expected header set to val, got %q", req.Header.Get("X-Test"))
	}
}

func TestRequestBuilderInvalidURL(t *testing.T) {
	b := NewRequestBuilder()
	_, err := b.Build(context.Background(), http.MethodGet, ":", nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}
