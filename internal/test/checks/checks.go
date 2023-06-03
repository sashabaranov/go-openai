package checks

import (
	"errors"
	"strings"
	"testing"
)

func NoError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err != nil {
		t.Error(err, message)
	}
}

func HasError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err == nil {
		t.Error(err, message)
	}
}

func ErrorIs(t *testing.T, err, target error, msg ...string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatal(msg)
	}
}

func ErrorIsF(t *testing.T, err, target error, format string, msg ...string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf(format, msg)
	}
}

func ErrorIsNot(t *testing.T, err, target error, msg ...string) {
	t.Helper()
	if errors.Is(err, target) {
		t.Fatal(msg)
	}
}

func ErrorIsNotf(t *testing.T, err, target error, format string, msg ...string) {
	t.Helper()
	if errors.Is(err, target) {
		t.Fatalf(format, msg)
	}
}

func ErrorContains(t *testing.T, err error, search string, message ...string) {
	t.Helper()
	if err == nil || search == "" || !strings.Contains(err.Error(), search) {
		t.Error(err, message)
	}
}
