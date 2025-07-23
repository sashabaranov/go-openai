package openai

type chatCompletionRequestOptions struct {
	RequestBodyModifier RequestBodyModifier
	ExtraHeader         map[string]string
}

type ChatCompletionRequestOption func(*chatCompletionRequestOptions)

type RequestBodyModifier func(rawBody []byte) ([]byte, error)

func WithRequestBodyModifier(modifier RequestBodyModifier) ChatCompletionRequestOption {
	return func(opts *chatCompletionRequestOptions) {
		opts.RequestBodyModifier = modifier
	}
}

func WithExtraHeader(header map[string]string) ChatCompletionRequestOption {
	return func(opts *chatCompletionRequestOptions) {
		opts.ExtraHeader = header
	}
}
