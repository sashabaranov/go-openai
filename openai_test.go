package openai_test

import (
	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
)

func setupOpenAITestServer() (client *Client, server *test.ServerTest, teardown func()) {
	server = test.NewTestServer()
	ts := server.OpenAITestServer()
	ts.Start()
	teardown = ts.Close
	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client = NewClientWithConfig(config)
	return
}

func setupAzureTestServer() (client *Client, server *test.ServerTest, teardown func()) {
	server = test.NewTestServer()
	ts := server.OpenAITestServer()
	ts.Start()
	teardown = ts.Close
	config := DefaultAzureConfig(test.GetTestToken(), "https://dummylab.openai.azure.com/")
	config.BaseURL = ts.URL
	client = NewClientWithConfig(config)
	return
}

func setupAzureTestServerWithCustomDeploymentName() (client *Client, server *test.ServerTest, teardown func()) {
	server = test.NewTestServer()
	ts := server.OpenAITestServer()
	ts.Start()
	teardown = ts.Close
	config := DefaultAzureConfig(test.GetTestToken(), "https://dummylab.openai.azure.com/")
	config.BaseURL = ts.URL
	config.AzureModelMapperFunc = func(model string) string {
		azureModelMapping := map[string]string{
			"gpt-3.5-turbo":      "custom-gpt-3.5-turbo",
			"gpt-3.5-turbo-0301": "custom-gpt-3.5-turbo-03-01",
			"gpt-4":              "custom-gpt-4",
			"gpt-4-0314":         "custom-gpt-4-03-14",
			"gpt-4-32k":          "custom-gpt-4-32k",
			"gpt-4-32k-0314":     "custom-gpt-4-32k-03-14",
		}
		return azureModelMapping[model]
	}
	client = NewClientWithConfig(config)
	return
}
