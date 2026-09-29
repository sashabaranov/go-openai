# September 2026 OpenAI API guide

Checked against the [OpenAI API changelog](https://developers.openai.com/api/docs/changelog)
on September 29, 2026. This SDK uses the published model IDs, including `gpt-6.1-sol`.
Model access and supported parameters are determined by the API and your project.

## Choose a model and endpoint

| Model | Go constant | Reasoning effort | Chat Completions tools |
| --- | --- | --- | --- |
| GPT-6.1 Sol | `GPT6Dot1Sol` | low, medium (default), high, xhigh, max | Use Responses |
| GPT-6 Astra | `GPT6Astra` | low, medium (default), high, xhigh, max | Use Responses |
| GPT-6 Sol | `GPT6Sol` | none, low, medium (default), high, xhigh, max | Only with explicit `none` |
| GPT-6 Luna | `GPT6Luna` | none, low, medium (default), high, xhigh, max | Only with explicit `none` |

GPT-6.1 Sol is a starting point for complex coding and professional work; Astra is
the most capable tier, while Luna suits focused, high-volume work. GPT-5.6 constants
(`GPT5Dot6`, `GPT5Dot6Sol`, `GPT5Dot6Terra`, `GPT5Dot6Luna`) are also available.
Strings remain accepted so model access does not depend on a library release.
See [GPT-6 guidance](https://developers.openai.com/api/docs/guides/latest-model)
and the [GPT-6.1 Sol model page](https://developers.openai.com/api/docs/models/gpt-6.1-sol).

Use `CreateResponse` or `CreateResponseStream` for reasoning with tools. For GPT-6
Chat Completions, omit `Temperature`, `TopP`, `LogProbs`, and `TopLogProbs` while
reasoning is enabled. Use `MaxCompletionTokens`, not `MaxTokens`. Sol 6.1 and Astra
reject `none` and `minimal`; Sol 6 and Luna support `none` but reject `minimal`.
Both chat methods perform the same checks for the four published GPT-6 model IDs.
Custom deployment names are validated by the server. The Responses API passes
request configuration through to the server without guessing model capabilities.

## Responses and streaming

The [conversation example](../examples/responses) uses `previous_response_id` to
continue a response. The [streaming example](../examples/responses-streaming)
consumes `response.output_text.delta` and checks the terminal response state.
`CreateResponseStream` uses HTTP server-sent events (SSE).

`GetOutputText()` retains its existing behavior: it aggregates output text,
including commentary and subagent messages. Use `GetFinalOutputText()` to display
only root-agent final answers. It filters items by `type`, `agent.agent_name`, and
`phase`; it accepts ordinary messages without agent or phase metadata. It reads
`Output` rather than the potentially aggregated `OutputText` field.

## Async function tools

```go
tool := openai.NewResponseAsyncFunctionTool(openai.FunctionDefinition{
    Name: "lookup",
    Description: "Look up a reference by ID.",
    Strict: true,
    Parameters: map[string]any{
        "type": "object",
        "properties": map[string]any{"id": map[string]any{"type": "string"}},
        "required": []string{"id"},
        "additionalProperties": false,
    },
})
request := openai.CreateResponseRequest{
    Model: openai.GPT6Dot1Sol,
    Input: "Look up reference 123 and explain the general process while it runs.",
    Tools: []openai.ResponseTool{tool},
}
```

The helper adds `async: true` to the inline Responses function schema. Your
application executes the function and returns a `ResponseFunctionCallOutput`
with the original `CallID` in a subsequent request. Keep track of the latest
response ID for continuations. Async tools do not run application code at OpenAI.
`ResponseOutputItem.Async` distinguishes async calls; streaming events retain
all fields in `Raw`. Custom tools can set `Parameters["async"] = true`.
See [async tool calling](https://developers.openai.com/api/docs/guides/async-tool-calling).

## Change reasoning between turns

Keep the initial request-level effort unchanged, and insert a configuration item
before the next user message:

```go
request := openai.CreateResponseRequest{
    Model: openai.GPT6Dot1Sol,
    PreviousResponseID: previous.ID,
    Reasoning: &openai.ResponseReasoning{Effort: openai.ReasoningEffortLow},
    Input: []any{
        openai.ResponseConfigurationUpdate{
            Type: "configuration_update",
            Reasoning: openai.ResponseReasoning{Effort: openai.ReasoningEffortHigh},
        },
        openai.ResponseInputMessage{Role: "user", Content: "Analyze the failure modes."},
    },
}
```

Updates persist until overridden. They work in standard single-agent mode, and
cannot be adjacent or combined with automatic compaction, automatic truncation,
or `/responses/compact`. See [reasoning configuration updates](https://developers.openai.com/api/docs/guides/reasoning#change-reasoning-mid-conversation).

## Multi-agent beta

```go
request := openai.CreateResponseRequest{
    Model: openai.GPT6Dot1Sol,
    Input: "Compare the proposals using two subagents, then synthesize the findings.",
    Betas: []string{openai.ResponseBetaMultiAgent},
    MultiAgent: &openai.ResponseMultiAgent{
        Enabled: true,
        MaxConcurrentSubagents: 2,
    },
}
```

`Betas` becomes the `OpenAI-Beta` HTTP header, not a JSON field. Pass it on each
create/stream request and through `RetrieveResponseOptions` or
`ResponseInputItemsListOptions` when retrieving the conversation. The beta must
be explicitly enabled; ordinary requests do not send a beta header.

`ResponseOutputItem.Agent` and `ResponseStreamEvent.Agent` preserve attribution.
Hosted `multi_agent_call` items are executed by OpenAI. Only execute your own
function calls; do not send outputs for hosted collaboration actions. Preserve
all output items when managing history manually, including encrypted messages.
`Output` remains `[]any` and stream events retain `Raw` for evolving schemas.

The beta supports GPT-6.1 Sol and GPT-5.6 models. The concurrent-subagent limit
excludes the root agent and defaults to three. More agents can increase token
usage. Multi-agent does not support `reasoning.summary`, `max_tool_calls`, or
`/responses/compact`; automatic compaction is implicit. See the
[official Multi-agent guide](https://developers.openai.com/api/docs/guides/responses-multi-agent)
and [runnable example](../examples/responses-multi-agent).

## Cache controls and service tiers

Existing `ResponsePromptCacheOptions` supports the current caching controls:

```go
request.PromptCacheOptions = &openai.ResponsePromptCacheOptions{TTL: "30m"}
```

For GPT-5.6 and later, use `prompt_cache_options.ttl` instead of
`prompt_cache_retention`. Inspect `Usage.InputTokensDetails.CachedTokens` and
`CacheWriteTokens` to measure reads and writes. See [prompt caching](https://developers.openai.com/api/docs/guides/prompt-caching).

Fast mode accepts `ServiceTierFast` (`fast`) or the existing `ServiceTierPriority`
(`priority`) value; see [Fast mode](https://developers.openai.com/api/docs/guides/fast-mode).
The new Ultrafast tier is
available for GPT-6 Astra (and GPT-5.6 Sol in limited preview):

```go
request.Model = openai.GPT6Astra
request.ServiceTier = string(openai.ServiceTierUltrafast)
```

Ultrafast supports global processing and US data residency, with separate pricing
and rate limits. It does not support EU or other regional processing endpoints.
See [Ultrafast mode](https://developers.openai.com/api/docs/guides/ultrafast-mode).

## GPT Image 2.5

Use `CreateImageModelGptImage2Dot5Sunburst` for precise edits or
`CreateImageModelGptImage2Dot5Flare` for everyday generation. Both accept `xhigh`
and `max` quality in addition to `low`, `medium`, `high`, and `auto`.

```go
request := openai.ImageRequest{
    Model: openai.CreateImageModelGptImage2Dot5Flare,
    Prompt: "A parrot on a skateboard, illustrated on a transparent background.",
    Size: openai.CreateImageSize1024x1024,
    Quality: openai.CreateImageQualityHigh,
    Background: openai.CreateImageBackgroundTransparent,
    OutputFormat: openai.CreateImageOutputFormatPNG,
}
```

GPT Image returns base64 data in `Data[i].B64JSON`; omit the legacy
`ResponseFormat` field. Decode and save the data as shown in the
[image example](../examples/images). Transparent backgrounds require PNG or WebP.
`ImageEditRequest` now sends model, quality, user, background, output format, and
optional output compression as multipart fields. Unset options are omitted so
API defaults apply; a compression pointer permits an explicit zero.
See [image generation](https://developers.openai.com/api/docs/guides/image-generation).

## Older endpoints and scope

The Assistants API shut down on August 26, 2026; migrate to Responses and, where
needed, the Conversations API. Some legacy model examples are retained as API
usage references. Consult [deprecations](https://developers.openai.com/api/docs/deprecations)
for shutdown dates and replacements.

This update covers the SDK's existing HTTP/SSE Responses, Chat Completions, and
Images surfaces. The separate Agents API, GPT-Live session protocol, and Responses
WebSocket transport (including mid-turn steering) do not yet have dedicated
clients here. `ExtraBody`, generic tools, and raw response items remain available
for additional fields on supported endpoints.
