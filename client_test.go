package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"io"
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
		name  string
		value interface{}
		body  io.Reader
	}{
		{
			name:  "nil input",
			value: nil,
			body:  bytes.NewReader([]byte("")),
		},
		{
			name:  "string input",
			value: &stringInput,
			body:  bytes.NewReader([]byte("test")),
		},
		{
			name:  "map input",
			value: &map[string]interface{}{},
			body:  bytes.NewReader([]byte(`{"test": "test"}`)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := decodeResponse(tc.body, tc.value)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
