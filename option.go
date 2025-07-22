package openai

type chatCompletionRequestOptions struct {
	RequestBodySetter RequestBodySetter
}

type ChatCompletionRequestOption func(*chatCompletionRequestOptions)

type RequestBodySetter func(rawBody []byte) ([]byte, error)

func WithRequestBodySetter(setter RequestBodySetter) ChatCompletionRequestOption {
	return func(opts *chatCompletionRequestOptions) {
		opts.RequestBodySetter = setter
	}
}
