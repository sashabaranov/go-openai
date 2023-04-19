package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestClient(t *testing.T) {
	const mockToken = "mock token"
	client := NewClient(mockToken)
	if client.config.authToken != mockToken {
		t.Errorf("Client does not contain proper token")
	}

	const mockOrg = "mock org"
	client = NewOrgClient(mockToken, mockOrg)
	if client.config.authToken != mockToken {
		t.Errorf("Client does not contain proper token")
	}
	if client.config.OrgID != mockOrg {
		t.Errorf("Client does not contain proper orgID")
	}
}

func TestDecodeResponse(t *testing.T) {
	stringInput := ""

	testCases := []struct {
		name     string
		input    interface{}
		response http.Response
	}{
		{
			name:  "nil input",
			input: nil,
			response: http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
		},
		{
			name:  "string input",
			input: &stringInput,
			response: http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("test"))),
			},
		},
		{
			name:  "map input",
			input: &map[string]interface{}{},
			response: http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"test": "test"}`))),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := decodeResponse(tc.input, tc.response)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
