package expecthttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cga1123/bissy-api/expect"
)

func Status(t *testing.T, expected int, rr *httptest.ResponseRecorder) {
	t.Helper()

	expect.Equal(t, expected, rr.Code)
}

type hasHeaders interface {
	Get(string) string
}

func Header(t *testing.T, key string, val string, headers hasHeaders) {
	t.Helper()

	expect.Equal(t, val, headers.Get(key))
}

func Ok(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	Status(t, http.StatusOK, rr)
}

func ContentType(t *testing.T, expected string, rr *httptest.ResponseRecorder) {
	t.Helper()

	expect.Equal(t, expected, rr.Header().Get("Content-Type"))
}

func StringBody(t *testing.T, expected string, rr *httptest.ResponseRecorder) {
	t.Helper()

	expect.Equal(t, expected, rr.Body.String())
}

type hasBytes interface {
	Bytes() []byte
}

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
