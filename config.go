package openai

import (
	"fmt"
	"net/http"
)

const (
	openaiApiURLv1                 = "https://api.openai.com/v1"
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

var supportedApiType = map[ApiType]struct{}{
	ApiTypeOpenAI:  {},
	ApiTypeAzure:   {},
	ApiTypeAzureAD: {},
}

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
		BaseURL:    openaiApiURLv1,
		OrgID:      "",
		authToken:  authToken,

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
}

func NewConfig(authTokenOrKey string, opts ...Option) (ClientConfig, error) {
	cfg := ClientConfig{
		ApiType:    ApiTypeOpenAI,
		Engine:     "",
		ApiVersion: "",
		HTTPClient: &http.Client{},
		BaseURL:    openaiApiURLv1,
		OrgID:      "",
		authToken:  authTokenOrKey,

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
	for _, o := range opts {
		o(&cfg)
	}
	if authTokenOrKey == "" {
		return ClientConfig{}, fmt.Errorf("auth token or key is required")
	}

	if _, ok := supportedApiType[cfg.ApiType]; !ok {
		return ClientConfig{}, fmt.Errorf("unsupported API type %s", cfg.ApiType)
	}

	if cfg.ApiType == ApiTypeAzure || cfg.ApiType == ApiTypeAzureAD {
		if cfg.ApiVersion == "" {
			return ClientConfig{}, fmt.Errorf("an API version is required for the Azure API type")
		}
	}

	return cfg, nil
}

type Option func(*ClientConfig)

// WithApiType sets the API type to use.
func WithApiType(apiType ApiType) Option {
	return func(o *ClientConfig) {
		o.ApiType = apiType
	}
}

// WithEngine sets the engine to use.
func WithEngine(engine string) Option {
	return func(o *ClientConfig) {
		o.Engine = engine
	}
}

// WithApiVersion sets the API version to use.
func WithApiVersion(apiVersion string) Option {
	return func(o *ClientConfig) {
		o.ApiVersion = apiVersion
	}
}

// WithHTTPClient sets the HTTP client to use.
func WithHTTPClient(client *http.Client) Option {
	return func(o *ClientConfig) {
		o.HTTPClient = client
	}
}

func WithBaseURL(apiBase string) Option {
	return func(o *ClientConfig) {
		o.BaseURL = apiBase
	}
}

// WithOrgID sets the organization ID to use.
func WithOrgID(orgID string) Option {
	return func(o *ClientConfig) {
		o.OrgID = orgID
	}
}
