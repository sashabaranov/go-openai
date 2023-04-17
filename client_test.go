package openai //nolint:testpackage // testing private field

import (
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

// test update auth token.
func TestUpdateAuthToken(t *testing.T) {
	const mockToken = "mock token"
	client := NewClient(mockToken)
	if client.config.authToken != mockToken {
		t.Errorf("Client does not contain proper token")
	}
	const updateToken = "update token"
	client.SetAuthToken(updateToken)
	if client.config.authToken != updateToken {
		t.Errorf("Client does not contain proper token")
	}
}

// test update org id.
func TestUpdateOrgID(t *testing.T) {
	const mockOrg = "mock org"
	client := NewOrgClient("mock token", mockOrg)
	if client.config.OrgID != mockOrg {
		t.Errorf("Client does not contain proper orgID")
	}
	const updateOrg = "update org"
	client.SetOrgID(updateOrg)
	if client.config.OrgID != updateOrg {
		t.Errorf("Client does not contain proper orgID")
	}
}

func TestClient_SetHTTPClient(t *testing.T) {
	client := NewClientWithConfig(ClientConfig{
		HTTPClient: &http.Client{},
	})
	updateHttpClient := &http.Client{}
	client.SetHTTPClient(updateHttpClient)
}

func TestClient_SetConfig(t *testing.T) {
	const configToken = "config token"
	const configOrg = "config org"
	const configBaseUrl = "config base url"

	mockConfig := ClientConfig{
		authToken:          configToken,
		BaseURL:            configBaseUrl,
		OrgID:              configOrg,
		APIType:            "",
		APIVersion:         "",
		Engine:             "",
		HTTPClient:         nil,
		EmptyMessagesLimit: 0,
	}
	client := NewClient("mock token")
	client.SetConfig(mockConfig)
	if client.config.authToken != configToken {
		t.Errorf("Client does not contain proper token")
	}
	if client.config.OrgID != configOrg {
		t.Errorf("Client does not contain proper orgID")
	}
	if client.config.BaseURL != configBaseUrl {
		t.Errorf("Client does not contain proper base url")
	}
}
