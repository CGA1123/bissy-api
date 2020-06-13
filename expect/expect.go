package expect

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// True expected the given value to be true
func True(t *testing.T, actual bool) {
	t.Helper()

	Equal(t, true, actual)
}

// False expected the given value to be false
func False(t *testing.T, actual bool) {
	t.Helper()

	Equal(t, false, actual)
}

// Equal expects the given values to be equal
func Equal(t *testing.T, expected, actual interface{}) {
	t.Helper()

	diff := cmp.Diff(expected, actual)
	if diff != "" {
		t.Errorf("Not Equal: %v", diff)
	}
}

// NotEqual expects the given values not to be equal
func NotEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()

	diff := cmp.Diff(expected, actual)
	if diff == "" {
		t.Errorf("Expected %v not to equal %v", expected, actual)
	}
}

// Ok expects the given error to be nil.
// Will fatally end the current test run if not.
func Ok(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		return
	}

	t.Fatalf("unexpected error: %v", err)
}

// Error expects the given error to be present.
// Will fatally end the current rest run if not.
func Error(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		return
	}

	t.Fatal("Expected error did not occur!")
}
