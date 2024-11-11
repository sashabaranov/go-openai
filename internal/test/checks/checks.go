package checks

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func NoError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err != nil {
		t.Error(err, message)
	}
}

func NoErrorF(t *testing.T, err error, message ...string) {
	t.Helper()
	if err != nil {
		t.Fatal(err, message)
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

type TestingT interface {
	Fatalf(format string, args ...any)
	Errorf(format string, args ...any)
}

type tHelper interface {
	Helper()
}

// Equal asserts that two objects are equal.
//
//	assert.Equal(t, 123, 123)
//
// Pointer variable equality is determined based on the equality of the
// referenced values (as opposed to the memory addresses). Function equality
// cannot be determined and will always fail.
func Equal(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if err := validateEqualArgs(expected, actual); err != nil {
		t.Fatalf("Invalid operation: %#v == %#v (%s)", expected, actual, err)
	}

	if !ObjectsAreEqual(expected, actual) {
		t.Fatalf("Not equal: \n"+
			"expected: %+v\n"+
			"actual  : %+v", expected, actual)
	}

	return true
}

// JSONEq asserts that two JSON strings are equivalent.
//
//	assert.JSONEq(t, `{"hello": "world", "foo": "bar"}`, `{"foo": "bar", "hello": "world"}`)
func JSONEq(t TestingT, expected string, actual string, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	var expectedJSONAsInterface, actualJSONAsInterface interface{}

	if err := json.Unmarshal([]byte(expected), &expectedJSONAsInterface); err != nil {
		t.Fatalf("Expected value ('%s') is not valid json.\nJSON parsing error: '%s'", expected, err.Error())
	}

	if err := json.Unmarshal([]byte(actual), &actualJSONAsInterface); err != nil {
		t.Fatalf("Input ('%s') needs to be valid json.\nJSON parsing error: '%s'", actual, err.Error())
	}

	return Equal(t, expectedJSONAsInterface, actualJSONAsInterface, msgAndArgs...)
}

// validateEqualArgs checks whether provided arguments can be safely used in the
// Equal/NotEqual functions.
func validateEqualArgs(expected, actual interface{}) error {
	if expected == nil && actual == nil {
		return nil
	}

	if isFunction(expected) || isFunction(actual) {
		return errors.New("cannot take func type as argument")
	}
	return nil
}

func isFunction(arg interface{}) bool {
	if arg == nil {
		return false
	}
	return reflect.TypeOf(arg).Kind() == reflect.Func
}

/*
	Helper functions
*/

// ObjectsAreEqual determines if two objects are considered equal.
//
// This function does no assertion of any kind.
func ObjectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}
