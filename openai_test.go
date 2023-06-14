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
