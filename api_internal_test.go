package openai

import (
	"testing"
)

func TestAzureFullURL(t *testing.T) {
	az := DefaultAzure("dummy", "https://httpbin.org/", "chatgpt-demo")
	cli := NewClientWithConfig(az)
	// /openai/deployments/{engine}/chat/completions?api-version={api_version}
	expect := "https://httpbin.org/" +
		"openai/deployments/chatgpt-demo/chat/completions?api-version=2023-03-15-preview"
	actual := cli.fullURL("/chat/completions")
	if actual != expect {
		t.Errorf("Expected %s, got %s", expect, actual)
	}
	t.Logf("Full URL: %s", actual)
}
