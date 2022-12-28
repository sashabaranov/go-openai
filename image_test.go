package gogpt_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	. "github.com/sashabaranov/go-gpt3"
)

func TestClient_CreateImageCreate(t *testing.T) {

	for _, tt := range []struct {
		name    string
		req     ImageCreateRequest
		expResp *ImageCreateResponse
		handler func(w http.ResponseWriter, r *http.Request)
		expErr  error
	}{
		{
			name: "prompt too long",
			req: ImageCreateRequest{
				Prompt: strings.Repeat("a", ImageMaxPromptLength+1),
			},
			expErr: fmt.Errorf("prompt too long, max length is %d", ImageMaxPromptLength),
		},
		{
			name: "success",
			req: ImageCreateRequest{
				Prompt:         "test",
				N:              5,
				Size:           ImageSizeBig,
				ResponseFormat: ImageResponseFormatURLs,
				User:           "user",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("unexpected method: %s", r.Method)
				}

				if r.URL.Path != URLImageGeneration {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}

				if r.Header.Get("Authorization") != "Bearer "+testAPIToken {
					t.Errorf("unexpected authorization header: %s", r.Header)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read body: %v", err)
				}

				var req ImageCreateRequest
				if err := json.Unmarshal(body, &req); err != nil {
					t.Errorf("unmarshal body: %v", err)
				}

				expReq := ImageCreateRequest{
					Prompt:         "test",
					N:              5,
					Size:           ImageSizeBig,
					ResponseFormat: ImageResponseFormatURLs,
					User:           "user",
				}
				if !reflect.DeepEqual(req, expReq) {
					t.Errorf("unexpected request: got: %v, want: %v", req, expReq)
				}

				_, err = w.Write([]byte(`{"created": 1589478378, "data": [{"url": "url1"}, {"url": "url2"}, {"b64_json": "b64_json1"}, {"b64_json": "b64_json2"}]}`))
				if err != nil {
					t.Errorf("write response: %v", err)
				}
			},
			expResp: &ImageCreateResponse{
				CreatedAt: time.UnixMilli(1589478378),
				URLs:      []string{"url1", "url2"},
				Images:    []string{"b64_json1", "b64_json2"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			client := NewClient(testAPIToken)
			client.BaseURL = server.URL

			gotResp, err := client.CreateImageCreate(context.Background(), tt.req)

			if tt.expErr != nil {
				if err.Error() != tt.expErr.Error() {
					t.Errorf("got error %v, want %v", err, tt.expErr)
				}
				return
			}

			if tt.expResp.CreatedAt.Unix() != gotResp.CreatedAt.Unix() {
				t.Errorf("got created at %v, want %v", gotResp.CreatedAt, tt.expResp.CreatedAt)
			}

			if tt.expResp.URLs != nil && !reflect.DeepEqual(gotResp.URLs, tt.expResp.URLs) {
				t.Errorf("got urls %v, want %v", gotResp.URLs, tt.expResp.URLs)
			}

			if tt.expResp.Images != nil && !reflect.DeepEqual(gotResp.Images, tt.expResp.Images) {
				t.Errorf("got images %v, want %v", gotResp.Images, tt.expResp.Images)
			}
		})
	}
}
