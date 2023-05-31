package test

import "errors"

var (
	ErrTestErrorAccumulatorWriteFailed = errors.New("test error accumulator failed")
)

type FailingErrorBuffer struct{}

func (b *FailingErrorBuffer) Write(_ []byte) (n int, err error) {
	return 0, ErrTestErrorAccumulatorWriteFailed
}

func (b *FailingErrorBuffer) Len() int {
	return 0
}

func (b *FailingErrorBuffer) Bytes() []byte {
	return []byte{}
}
