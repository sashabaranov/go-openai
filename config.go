package gogpt

import (
	"net/http"
)

const apiURLv1 = "https://api.openai.com/v1"

type ClientConfig struct {
	HTTPClient *http.Client

	BaseURL   string
	OrgID     string
	AuthToken string
}

func DefaultConfig(authToken string) ClientConfig {
	return ClientConfig{
		HTTPClient: &http.Client{},
		BaseURL:    apiURLv1,
		OrgID:      "",
		AuthToken:  authToken,
	}
}
