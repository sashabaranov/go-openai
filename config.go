package openai

import (
	"fmt"
	"net/http"
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

var supportedAPIType = map[APIType]struct{}{
	APITypeOpenAI:  {},
	APITypeAzure:   {},
	APITypeAzureAD: {},
}

// ClientConfig is a configuration of a client.
type ClientConfig struct {
	APIType    APIType
	APIKey     string
	APIBase    string
	APIVersion string

	Engine string
	OrgID  string

	HTTPClient *http.Client

	EmptyMessagesLimit uint
}

func DefaultConfig(apiKey string) (ClientConfig, error) {
	return NewConfig(WithAPIKey(apiKey))
}

func NewConfig(opts ...Option) (ClientConfig, error) {
	cfg := ClientConfig{
		APIType:            APITypeOpenAI,
		APIKey:             "",
		APIBase:            openaiAPIURLv1,
		APIVersion:         "",
		Engine:             "",
		OrgID:              "",
		HTTPClient:         &http.Client{},
		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.APIKey == "" {
		return ClientConfig{}, fmt.Errorf("api key is required")
	}

	if _, ok := supportedAPIType[cfg.APIType]; !ok {
		return ClientConfig{}, fmt.Errorf("unsupported API type %s", cfg.APIType)
	}

	if cfg.APIType == APITypeAzure || cfg.APIType == APITypeAzureAD {
		if cfg.APIVersion == "" {
			return ClientConfig{}, fmt.Errorf("an API version is required for the Azure API type")
		}
	}

	return cfg, nil
}

type Option func(*ClientConfig)

// WithAPIType sets the API type to use.
func WithAPIType(apiType APIType) Option {
	return func(o *ClientConfig) {
		o.APIType = apiType
	}
}

// WithEngine sets the engine to use.
func WithEngine(engine string) Option {
	return func(o *ClientConfig) {
		o.Engine = engine
	}
}

// WithAPIVersion sets the API version to use.
func WithAPIVersion(apiVersion string) Option {
	return func(o *ClientConfig) {
		o.APIVersion = apiVersion
	}
}

// WithHTTPClient sets the HTTP client to use.
func WithHTTPClient(client *http.Client) Option {
	return func(o *ClientConfig) {
		o.HTTPClient = client
	}
}

func WithAPIBase(apiBase string) Option {
	return func(o *ClientConfig) {
		o.APIBase = apiBase
	}
}

func WithAPIKey(apiKey string) Option {
	return func(o *ClientConfig) {
		o.APIKey = apiKey
	}
}

// WithOrgID sets the organization ID to use.
func WithOrgID(orgID string) Option {
	return func(o *ClientConfig) {
		o.OrgID = orgID
	}
}
