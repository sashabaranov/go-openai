# Responses API example

This example shows how to create a response, read its text, and continue the
conversation with `previous_response_id`.

```sh
export OPENAI_API_KEY="<your key here>"
go run ./examples/responses
```

Set `OPENAI_MODEL` to use a model other than `gpt-5.6-sol`.

## Application-handled documentation lookup

The optional docs mode shows an existing Responses function call executed by the
application, followed by a `function_call_output` carrying the same `call_id`:

```sh
go run ./examples/responses -docs
```

It asks why Go code should defer cancellation after `context.WithTimeout`. The
application supplies a `find_go_docs` callback that searches public documentation
through Parallel's Unified MCP. It selects an HTTPS result on `go.dev` or
`pkg.go.dev` with an excerpt, passes that evidence back to the Responses API, and
checks that the answer cites the exact returned URL. URL validation checks the
source domain; it does not verify every claim in the generated answer.

The normal example above keeps its existing behavior. Docs mode is opt-in and
requires `OPENAI_API_KEY` for the two Responses requests, which may incur OpenAI
charges. The public documentation question is sent to both OpenAI and Parallel;
use public inputs only. The Parallel callback uses anonymous, rate-limited access
with no Parallel key and a separate HTTP client that receives no OpenAI
credentials. See [Parallel MCP setup](https://docs.parallel.ai/integrations/mcp/quickstart)
for access options and service terms. The callback stops on missing sources,
HTTP or tool errors, cancellation, or timeout; it does not retry or switch to
paid access.

This is application code using the SDK's existing native function types. The
small callback implements only the JSON/SSE MCP exchange needed by this example;
it is not a general MCP client. `docsLookup` can be replaced with another public
documentation source. See the [OpenAI function calling guide](https://developers.openai.com/api/docs/guides/function-calling)
for the API contract.

### Validation

```sh
go test ./examples/responses
GO_OPENAI_LIVE_DOCS=1 go test -v ./examples/responses -run '^TestLiveParallelDocs$' -count=1
```

The default tests run the actual SDK create-response and tool-output requests
against `httptest` servers with raw JSON fixtures for both model turns. They
check inline function encoding, callback execution, the matched call ID, exact
source citation, credential separation, absent results, and errors. These model
fixtures do not demonstrate live model behavior.

The separate opt-in test performs real anonymous MCP retrieval, checks the
example's User-Agent on the outgoing requests, and fetches the exact returned
Go documentation URL. It does not call OpenAI. Running `go run ./examples/responses -docs`
with your own OpenAI key is the live application round trip.
