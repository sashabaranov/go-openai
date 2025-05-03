package openai_test

import (
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestGetAzureDeploymentByModel(t *testing.T) {
	cases := []struct {
		Model                string
		AzureModelMapperFunc func(model string) string
		Expect               string
	}{
		{
			Model:  "gpt-3.5-turbo",
			Expect: "gpt-35-turbo",
		},
		{
			Model:  "gpt-3.5-turbo-0301",
			Expect: "gpt-35-turbo-0301",
		},
		{
			Model:  "text-embedding-ada-002",
			Expect: "text-embedding-ada-002",
		},
		{
			Model:  "",
			Expect: "",
		},
		{
			Model:  "models",
			Expect: "models",
		},
		{
			Model:  "gpt-3.5-turbo",
			Expect: "my-gpt35",
			AzureModelMapperFunc: func(model string) string {
				modelmapper := map[string]string{
					"gpt-3.5-turbo": "my-gpt35",
				}
				if val, ok := modelmapper[model]; ok {
					return val
				}
				return model
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Model, func(t *testing.T) {
			conf := openai.DefaultAzureConfig("", "https://test.openai.azure.com/")
			if c.AzureModelMapperFunc != nil {
				conf.AzureModelMapperFunc = c.AzureModelMapperFunc
			}
			actual := conf.GetAzureDeploymentByModel(c.Model)
			if actual != c.Expect {
				t.Errorf("Expected %s, got %s", c.Expect, actual)
			}
		})
	}
}

func TestDefaultAnthropicConfig(t *testing.T) {
	apiKey := "test-key"
	baseURL := "https://api.anthropic.com/v1"

	config := openai.DefaultAnthropicConfig(apiKey, baseURL)

	if config.APIType != openai.APITypeAnthropic {
		t.Errorf("Expected APIType to be %v, got %v", openai.APITypeAnthropic, config.APIType)
	}

	if config.APIVersion != openai.AnthropicAPIVersion {
		t.Errorf("Expected APIVersion to be 2023-06-01, got %v", config.APIVersion)
	}

	if config.BaseURL != baseURL {
		t.Errorf("Expected BaseURL to be %v, got %v", baseURL, config.BaseURL)
	}

	if config.EmptyMessagesLimit != 300 {
		t.Errorf("Expected EmptyMessagesLimit to be 300, got %v", config.EmptyMessagesLimit)
	}
}

func TestDefaultAnthropicConfigWithEmptyValues(t *testing.T) {
	config := openai.DefaultAnthropicConfig("", "")

	if config.APIType != openai.APITypeAnthropic {
		t.Errorf("Expected APIType to be %v, got %v", openai.APITypeAnthropic, config.APIType)
	}

	if config.APIVersion != openai.AnthropicAPIVersion {
		t.Errorf("Expected APIVersion to be %s, got %v", openai.AnthropicAPIVersion, config.APIVersion)
	}

	expectedBaseURL := "https://api.anthropic.com/v1"
	if config.BaseURL != expectedBaseURL {
		t.Errorf("Expected BaseURL to be %v, got %v", expectedBaseURL, config.BaseURL)
	}
}

func TestClientConfigString(t *testing.T) {
	// String() should always return the constant value
	conf := openai.DefaultConfig("dummy-token")
	expected := "<OpenAI API ClientConfig>"
	got := conf.String()
	if got != expected {
		t.Errorf("ClientConfig.String() = %q; want %q", got, expected)
	}
}

func TestGetAzureDeploymentByModel_NoMapper(t *testing.T) {
	// On a zero-value or DefaultConfig, AzureModelMapperFunc is nil,
	// so GetAzureDeploymentByModel should just return the input model.
	conf := openai.DefaultConfig("dummy-token")
	model := "some-model"
	got := conf.GetAzureDeploymentByModel(model)
	if got != model {
		t.Errorf("GetAzureDeploymentByModel(%q) = %q; want %q", model, got, model)
	}
}
