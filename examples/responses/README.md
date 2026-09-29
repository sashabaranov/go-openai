# Responses API examples

Create a GPT-6.1 Sol response and continue with `previous_response_id`:

```sh
export OPENAI_API_KEY="<your key here>"
go run ./examples/responses
```

Set `OPENAI_MODEL` to override `gpt-6.1-sol`. The example explicitly uses `low`
reasoning; choose a model supporting that effort.

Stream response text with terminal-state checks:

```sh
go run ./examples/responses-streaming
```

Try hosted Multi-agent orchestration with GPT-6.1 Sol (beta):

```sh
go run ./examples/responses-multi-agent
```

The Multi-agent example explicitly sends the beta header and displays only the
root agent's final answer. It may use more tokens than a single-agent request.
