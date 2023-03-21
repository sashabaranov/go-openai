package checks

import (
	"testing"
)

func NoError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err != nil {
		t.Error(err, message)
	}
}
