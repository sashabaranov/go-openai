# Responses API example

This example exercises a broad Responses API workflow:

- input-token counting before generation;
- structured message input and strict JSON-schema output;
- a built-in web-search tool and a local function tool;
- parallel tool calls and a multi-round `function_call_output` loop;
- conversation chaining with `previous_response_id`;
- response retrieval, input-item listing, usage data, and cleanup;
- a streamed follow-up; and
- optional background polling/cancellation and standalone compaction.

Set an API key and run the core workflow:

```sh
export OPENAI_API_KEY="<your key here>"
go run ./examples/responses
```

Run every optional section:

```sh
go run ./examples/responses -background -compact
```

`OPENAI_MODEL` or `-model` selects another model. The compaction flag requires
a model that supports the standalone compact endpoint. Stored responses are
deleted after the example; pass `-keep` to retain them for inspection.
