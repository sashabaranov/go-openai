package openai

type ConfigOption func(config *ClientConfig)

// WithBaseURL configures base url which should start with "https", e.g. https://exmample.com
func WithBaseURL(baseURL string) ConfigOption {
	return func(config *ClientConfig) {
		config.BaseURL = baseURL
	}
}

func WithAPIType(apiType APIType) ConfigOption {
	return func(config *ClientConfig) {
		config.APIType = apiType
	}
}

func WithAPIVersion(apiVersion string) ConfigOption {
	return func(config *ClientConfig) {
		config.APIVersion = apiVersion
	}
}

func WithOrgID(orgID string) ConfigOption {
	return func(config *ClientConfig) {
		config.OrgID = orgID
	}
}

func WithModelMapperFunc(mapper func(model string) string) ConfigOption {
	return func(config *ClientConfig) {
		config.AzureModelMapperFunc = mapper
	}
}

func WithEmptyMessagesLimit(limit uint) ConfigOption {
	return func(config *ClientConfig) {
		config.EmptyMessagesLimit = limit
	}
}
