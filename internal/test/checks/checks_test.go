package checks

import (
	"errors"
	"testing"
)

func TestChecksSuccessPaths(t *testing.T) {
	NoError(t, nil)
	NoErrorF(t, nil)
	HasError(t, errors.New("err"))
	target := errors.New("x")
	ErrorIs(t, target, target)
	ErrorIsF(t, target, target, "msg")
	ErrorIsNot(t, errors.New("y"), target)
	ErrorIsNotf(t, errors.New("y"), target, "msg")
}
