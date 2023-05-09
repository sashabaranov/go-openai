package openai_test

import (
	"github.com/sashabaranov/go-openai"
	"testing"
)

func TestGetAzureDeploymentByModel(t *testing.T) {
	cases := []struct {
		Model       string
		ModelMapper map[string]string
		Expect      string
	}{
		{
			Model:       "gpt-3.5-turbo",
			Expect:      "gpt-35-turbo",
			ModelMapper: nil,
		},
		{
			Model:       "gpt-3.5-turbo-0301",
			Expect:      "gpt-35-turbo-0301",
			ModelMapper: nil,
		},
		{
			Model:       "text-embedding-ada-002",
			Expect:      "text-embedding-ada-002",
			ModelMapper: nil,
		},
		{
			Model:       "",
			Expect:      "",
			ModelMapper: nil,
		},
		{
			Model:       "models",
			Expect:      "models",
			ModelMapper: nil,
		},
		{
			Model:  "gpt-3.5-turbo",
			Expect: "my-gpt35",
			ModelMapper: map[string]string{
				"gpt-3.5-turbo": "my-gpt35",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Model, func(t *testing.T) {
			conf := openai.DefaultAzureConfig("a", "https://test.openai.azure.com/", c.ModelMapper)
			actual := conf.GetAzureDeploymentByModel(c.Model)
			if actual != c.Expect {
				t.Errorf("Expected %s, got %s", c.Expect, actual)
			}
		})
	}
}
