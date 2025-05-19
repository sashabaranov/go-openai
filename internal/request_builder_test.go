package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type testExtraFieldsRequest struct {
	Model       string `json:"model"`
	extraFields map[string]any
}

func (r *testExtraFieldsRequest) GetExtraFields() map[string]any {
	return r.extraFields
}

func TestRequestBuilderReturnsRequestWhenRequestHasExtraFields(t *testing.T) {
	b := NewRequestBuilder()
	var (
		ctx     = context.Background()
		method  = http.MethodPost
		url     = "/foo"
		request = &testExtraFieldsRequest{
			Model: "test-model",
		}
	)
	request.extraFields = map[string]any{"extra_field": "extra_value"}

	reqBytes, err := b.marshaller.Marshal(request)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// 验证序列化结果包含原始字段和额外字段
	var result map[string]interface{}
	if err := json.Unmarshal(reqBytes, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result["model"] != "test-model" {
		t.Errorf("Expected model to be 'test-model', got %v", result["model"])
	}
	if result["extra_field"] != "extra_value" {
		t.Errorf("Expected extra_field to be 'extra_value', got %v", result["extra_field"])
	}

	want, _ := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBytes))
	got, _ := b.Build(ctx, method, url, request, nil)
	if !reflect.DeepEqual(got.Body, want.Body) ||
		!reflect.DeepEqual(got.URL, want.URL) ||
		!reflect.DeepEqual(got.Method, want.Method) {
		t.Errorf("Build() got = %v, want %v", got, want)
	}
}
