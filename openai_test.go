package openai_test

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
)

func setupOpenAITestServer() (client *openai.Client, server *test.ServerTest, teardown func()) {
	return setupOpenAITestServerWithAuth(test.GetTestToken())
}

func setupOpenAITestServerWithAuth(authKey string) (client *openai.Client, server *test.ServerTest, teardown func()) {
	server = test.NewTestServer(authKey)
	ts := server.OpenAITestServer()
	ts.Start()
	teardown = ts.Close
	config := openai.DefaultConfig(authKey)
	config.BaseURL = ts.URL + "/v1"
	config.DashboardBaseURL = ts.URL + "/dashboard"
	client = openai.NewClientWithConfig(config)
	return
}

func setupAzureTestServer() (client *openai.Client, server *test.ServerTest, teardown func()) {
	server = test.NewTestServer(test.GetTestToken())
	ts := server.OpenAITestServer()
	ts.Start()
	teardown = ts.Close
	config := openai.DefaultAzureConfig(test.GetTestToken(), "https://dummylab.openai.azure.com/")
	config.BaseURL = ts.URL
	client = openai.NewClientWithConfig(config)
	return
}

// numTokens Returns the number of GPT-3 encoded tokens in the given text.
// This function approximates based on the rule of thumb stated by OpenAI:
// https://beta.openai.com/tokenizer
//
// TODO: implement an actual tokenizer for GPT-3 and Codex (once available)
func numTokens(s string) int {
	return int(float32(len(s)) / 4)
}
