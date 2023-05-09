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

	BaseURL          string
	OrgID            string
	APIType          APIType
	APIVersion       string            // required when APIType is APITypeAzure or APITypeAzureAD
	AzureModelMapper map[string]string // replace model to azure deployment name
	HTTPClient       *http.Client

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

func DefaultAzureConfig(apiKey, baseURL string, modelMapper map[string]string) ClientConfig {
	if modelMapper == nil || len(modelMapper) == 0 {
		modelMapper = map[string]string{
			GPT3Dot5Turbo0301: "gpt-35-turbo-0301",
			GPT3Dot5Turbo:     "gpt-35-turbo",
		}
	}

	return ClientConfig{
		authToken:        apiKey,
		BaseURL:          baseURL,
		OrgID:            "",
		APIType:          APITypeAzure,
		APIVersion:       "2023-03-15-preview",
		AzureModelMapper: modelMapper,

		HTTPClient: &http.Client{},

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
}

func (ClientConfig) String() string {
	return "<OpenAI API ClientConfig>"
}

func (c ClientConfig) GetAzureDeploymentByModel(model string) string {
	if v, ok := c.AzureModelMapper[model]; ok {
		return v
	}

	return regexp.MustCompile(`[.:]`).ReplaceAllString(model, "")
}
