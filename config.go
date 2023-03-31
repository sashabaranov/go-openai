package openai

import (
	"net/http"
)

const (
	apiURLv1                       = "https://api.openai.com/v1"
	defaultEmptyMessagesLimit uint = 300

	azureApiPrefix         = "openai"
	azureDeploymentsPrefix = "deployments"
)

type ApiType string

const (
	ApiTypeOpenAI  ApiType = "OPEN_AI"
	ApiTypeAzure   ApiType = "AZURE"
	ApiTypeAzureAD ApiType = "AZURE_AD"
)

// ClientConfig is a configuration of a client.
type ClientConfig struct {
	ApiType    ApiType
	Engine     string
	ApiVersion string

	authToken string

	HTTPClient *http.Client
	BaseURL    string
	OrgID      string

	EmptyMessagesLimit uint
}

func DefaultConfig(authToken string) ClientConfig {
	return ClientConfig{
		HTTPClient: &http.Client{},
		BaseURL:    apiURLv1,
		OrgID:      "",
		authToken:  authToken,

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
}

func DefaultAzureConfig(apiBase, engine, apiKey, apiVersion string) ClientConfig {
	return ClientConfig{
		ApiType:    ApiTypeAzure,
		Engine:     engine,
		ApiVersion: apiVersion,
		HTTPClient: &http.Client{},
		BaseURL:    apiBase,
		OrgID:      "",
		authToken:  apiKey,

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
}
