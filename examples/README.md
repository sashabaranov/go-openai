To run an example:

```sh
export OPENAI_API_KEY="<your key here>"
go run ./examples/<target>
```

Start with these current examples:

- [Responses conversation](responses): GPT-6.1 Sol with a follow-up using `previous_response_id`.
- [Responses streaming](responses-streaming): text deltas and terminal-state checks.
- [Responses Multi-agent beta](responses-multi-agent): hosted subagents and root final-answer extraction.
- [Images](images): GPT Image 2.5 Flare generation, saved to `image.png`.

Other directories contain older endpoint examples. Check model availability and
[OpenAI's deprecations](https://developers.openai.com/api/docs/deprecations) before
running them. These examples make real API calls and require an API key.
