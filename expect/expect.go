package expect

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Equal(t *testing.T, expected, actual interface{}) {
	diff := cmp.Diff(expected, actual)
	if diff != "" {
		t.Errorf("Not Equal: %v", diff)
	}
}

func Ok(t *testing.T, err error) {
	if err == nil {
		return
	}

	t.Fatalf("unexpected error: %v", err)
}

func Error(t *testing.T, err error) {
	if err != nil {
		return
	}

	t.Fatal("Expected error did not occur!")
}
