package expect

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func StatusHTTP(t *testing.T, expected int, rr *httptest.ResponseRecorder) {
	Equal(t, expected, rr.Code)
}

func StatusOK(t *testing.T, rr *httptest.ResponseRecorder) {
	StatusHTTP(t, http.StatusOK, rr)
}

func BodyString(t *testing.T, expected string, rr *httptest.ResponseRecorder) {
	Equal(t, expected, rr.Body.String())
}

func ContentType(t *testing.T, expected string, rr *httptest.ResponseRecorder) {
	Equal(t, expected, rr.Header().Get("Content-Type"))
}

func BodyJSON(t *testing.T, expected interface{}, rr *httptest.ResponseRecorder) {
	var body interface{}

	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Errorf("failed to parse body %v", err)
		return
	}

	Equal(t, expected, body)
}
