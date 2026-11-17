package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

const (
	parallelMCPURL = "https://mcp.parallel.ai/mcp"
	docsUserAgent  = "go-openai-responses-docs-example/1.0"
	mcpVersion     = "2025-03-26"
	maxMCPBody     = 2 << 20
)

// parallelDocs is an application-supplied callback for this example, not an SDK
// provider. It uses only anonymous web_search against Parallel's Unified MCP.
func parallelDocs(ctx context.Context, client *http.Client, endpoint, query string) (docSource, error) {
	mcp := docsMCP{client: client, endpoint: endpoint}
	var initialized struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if err := mcp.call(ctx, 1, "initialize", map[string]any{
		"protocolVersion": mcpVersion,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]string{"name": "go-openai-responses-docs-example", "version": "1.0"},
	}, &initialized); err != nil {
		return docSource{}, err
	}
	if initialized.ProtocolVersion != mcpVersion {
		return docSource{}, fmt.Errorf("unsupported MCP version %q", initialized.ProtocolVersion)
	}
	if err := mcp.call(ctx, 0, "notifications/initialized", nil, nil); err != nil {
		return docSource{}, err
	}
	var result struct {
		IsError bool `json:"isError"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StructuredContent json.RawMessage `json:"structuredContent"`
	}
	if err := mcp.call(ctx, 2, "tools/call", map[string]any{
		"name": "web_search",
		"arguments": map[string]any{
			"objective":      "Find official Go documentation answering this public question: " + query,
			"search_queries": []string{"site:go.dev context WithTimeout cancel", "site:pkg.go.dev context WithTimeout cancel"},
		},
	}, &result); err != nil {
		return docSource{}, err
	}
	if result.IsError {
		return docSource{}, errors.New("parallel web_search returned a tool error")
	}
	data := result.StructuredContent
	if len(data) == 0 || string(data) == "null" {
		for _, content := range result.Content {
			if content.Type == "text" {
				data = json.RawMessage(content.Text)
				break
			}
		}
	}
	return goDocResult(data)
}

func goDocResult(data []byte) (docSource, error) {
	var search struct {
		Results []struct {
			URL      string   `json:"url"`
			Excerpts []string `json:"excerpts"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &search); err != nil {
		return docSource{}, fmt.Errorf("decode search results: %w", err)
	}
	for _, result := range search.Results {
		excerpt := strings.Join(result.Excerpts, "\n")
		if officialGoURL(result.URL) && strings.TrimSpace(excerpt) != "" {
			return docSource{URL: result.URL, Excerpt: excerpt}, nil
		}
	}
	return docSource{}, errors.New("search returned no official HTTPS Go documentation with an excerpt")
}

func officialGoURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.User == nil && u.Port() == "" &&
		(u.Host == "go.dev" || u.Host == "pkg.go.dev")
}

// docsMCP implements only the bounded JSON/SSE request-response exchange needed
// by this callback. It is not a general-purpose MCP client (no OAuth or retries).
// No OpenAI client, credentials or environment variables are used here.
type docsMCP struct {
	client   *http.Client
	endpoint string
	session  string
}

func (m *docsMCP) call(ctx context.Context, id int, method string, params, result any) error {
	request := map[string]any{"jsonrpc": "2.0", "method": method}
	if params != nil {
		request["params"] = params
	}
	if id != 0 {
		request["id"] = id
	}
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("User-Agent", docsUserAgent)
	if m.session != "" {
		req.Header.Set("Mcp-Session-Id", m.session)
	}
	if method != "initialize" {
		req.Header.Set("MCP-Protocol-Version", mcpVersion)
	}
	response, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("MCP %s: HTTP %d", method, response.StatusCode)
	}
	if method == "initialize" {
		m.session = response.Header.Get("Mcp-Session-Id")
	}
	if id == 0 {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxMCPBody+1))
	if err != nil {
		return err
	}
	if len(data) > maxMCPBody {
		return errors.New("MCP response exceeds example size limit")
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	if mediaType == "text/event-stream" {
		data = mcpEventData(data)
	} else if mediaType != "application/json" {
		return fmt.Errorf("unexpected MCP content type %q", mediaType)
	}
	var envelope struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      int             `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err = json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("decode MCP response: %w", err)
	}
	if envelope.JSONRPC != "2.0" || envelope.ID != id {
		return errors.New("MCP response version or ID mismatch")
	}
	if envelope.Error != nil {
		return fmt.Errorf("MCP error %d: %s", envelope.Error.Code, envelope.Error.Message)
	}
	return json.Unmarshal(envelope.Result, result)
}

func mcpEventData(data []byte) []byte {
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), maxMCPBody)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			lines = append(lines, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	return []byte(strings.Join(lines, "\n"))
}
