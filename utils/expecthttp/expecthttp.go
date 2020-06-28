package expecthttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cga1123/bissy-api/utils/expect"
)

// Status expects the given status codes to match
func Status(t *testing.T, expected int, rr *httptest.ResponseRecorder) {
	t.Helper()

	expect.Equal(t, expected, rr.Code)
}

type hasHeaders interface {
	Get(string) string
}

// Header expects the given response to contain the given header
func Header(t *testing.T, key string, val string, headers hasHeaders) {
	t.Helper()

	expect.Equal(t, val, headers.Get(key))
}

// Ok expects the given response to have status code OK (200)
func Ok(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	Status(t, http.StatusOK, rr)
}

// ContentType expects the given response to have the given Content-Type header
func ContentType(t *testing.T, expected string, rr *httptest.ResponseRecorder) {
	t.Helper()

	expect.Equal(t, expected, rr.Header().Get("Content-Type"))
}

// StringBody expect the response body to match the given string
func StringBody(t *testing.T, expected string, rr *httptest.ResponseRecorder) {
	t.Helper()

	expect.Equal(t, expected, rr.Body.String())
}

type hasBytes interface {
	Bytes() []byte
}

// JSONBody expects the response body to match the marshaled struct given
func JSONBody(t *testing.T, expected interface{}, body hasBytes) {
	t.Helper()

	var actualBody, expectedBody interface{}

	expectedBytes, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("failed to marshal expected %v", err)
		return
	}

	if err := json.Unmarshal(expectedBytes, &expectedBody); err != nil {
		t.Errorf("failed to unmarshal exepcted %v", err)
		return
	}

	if err := json.Unmarshal(body.Bytes(), &actualBody); err != nil {
		t.Errorf("failed to unmarshal actual %v", err)
		return
	}

	expect.Equal(t, expectedBody, actualBody)
}
