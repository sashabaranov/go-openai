package openai

import "net/http"

type Option func(c *ClientConfig)

func WithCustomBaseURL(url string) Option {
	return func(c *ClientConfig) {
		c.BaseURL = url
	}
}

func WithOrganizationID(orgID string) Option {
	return func(c *ClientConfig) {
		c.OrgID = orgID
	}
}

func WithSpecificAPIType(apiType APIType) Option {
	return func(c *ClientConfig) {
		c.APIType = apiType
	}
}

func WithCustomAPIVersion(version string) Option {
	return func(c *ClientConfig) {
		c.APIVersion = version
	}
}

func WithCustomEngine(engine string) Option {
	return func(c *ClientConfig) {
		c.Engine = engine
	}
}

func WithCustomClient(client *http.Client) Option {
	return func(c *ClientConfig) {
		c.HTTPClient = client
	}
}

func WithEmptyMessagesLimit(limit uint) Option {
	return func(c *ClientConfig) {
		c.EmptyMessagesLimit = limit
	}
}
