package openai

import (
	"fmt"

	"github.com/tiktoken-go/tokenizer"
)

func Tokenize(model string, text string) (ids []uint, tokens []string, err error) {
	var c tokenizer.Codec
	c, err = tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		err = fmt.Errorf("model not supported: %w", err)
		return
	}

	return c.Encode(text)
}
