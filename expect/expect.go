package expect

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func True(t *testing.T, actual bool) {
	t.Helper()

	Equal(t, true, actual)
}

func False(t *testing.T, actual bool) {
	t.Helper()

	Equal(t, false, actual)
}

func Equal(t *testing.T, expected, actual interface{}) {
	t.Helper()

	diff := cmp.Diff(expected, actual)
	if diff != "" {
		t.Errorf("Not Equal: %v", diff)
	}
}

func Ok(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		return
	}

	t.Fatalf("unexpected error: %v", err)
}

func Error(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		return
	}

	t.Fatal("Expected error did not occur!")
}
