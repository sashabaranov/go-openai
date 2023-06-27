package openai

import (
	"net/http"
	"regexp"
)

const (
	openaiAPIURLv1                 = "https://api.openai.com/v1"
	defaultEmptyMessagesLimit uint = 300

	azureAPIPrefix         = "openai"
	azureDeploymentsPrefix = "deployments"
	azureDefaultAPIVersion = "2023-05-15"
)

type APIType string

const (
	APITypeOpenAI  APIType = "OPEN_AI"
	APITypeAzure   APIType = "AZURE"
	APITypeAzureAD APIType = "AZURE_AD"
)

const AzureAPIKeyHeader = "api-key"

// ClientConfig is a configuration of a client.
type ClientConfig struct {
	authToken string

	BaseURL              string
	OrgID                string
	APIType              APIType
	APIVersion           string                    // required when APIType is APITypeAzure or APITypeAzureAD
	AzureModelMapperFunc func(model string) string // replace model to azure deployment name func
	HTTPClient           *http.Client

	EmptyMessagesLimit uint
}

func DefaultConfig(authToken string) ClientConfig {
	return ClientConfig{
		authToken: authToken,
		BaseURL:   openaiAPIURLv1,
		APIType:   APITypeOpenAI,
		OrgID:     "",

		HTTPClient: &http.Client{},

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
}

type AzureConfigOption func(*ClientConfig)

func WithAzureAPIVersion(apiVersion string) AzureConfigOption {
	return func(c *ClientConfig) {
		c.APIVersion = apiVersion
	}
}

func DefaultAzureConfig(apiKey, baseURL string, opts ...AzureConfigOption) ClientConfig {
	c := ClientConfig{
		authToken:  apiKey,
		BaseURL:    baseURL,
		OrgID:      "",
		APIType:    APITypeAzure,
		APIVersion: azureDefaultAPIVersion,
		AzureModelMapperFunc: func(model string) string {
			return regexp.MustCompile(`[.:]`).ReplaceAllString(model, "")
		},

		HTTPClient: &http.Client{},

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

func (ClientConfig) String() string {
	return "<OpenAI API ClientConfig>"
}

func (c ClientConfig) GetAzureDeploymentByModel(model string) string {
	if c.AzureModelMapperFunc != nil {
		return c.AzureModelMapperFunc(model)
	}

	return model
}
