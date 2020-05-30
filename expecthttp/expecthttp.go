package expecthttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cga1123/bissy-api/expect"
)

func Status(t *testing.T, expected int, rr *httptest.ResponseRecorder) {
	expect.Equal(t, expected, rr.Code)
}

func Ok(t *testing.T, rr *httptest.ResponseRecorder) {
	Status(t, http.StatusOK, rr)
}

func ContentType(t *testing.T, expected string, rr *httptest.ResponseRecorder) {
	expect.Equal(t, expected, rr.Header().Get("Content-Type"))
}

func StringBody(t *testing.T, expected string, rr *httptest.ResponseRecorder) {
	expect.Equal(t, expected, rr.Body.String())
}

func JSONBody(t *testing.T, expected interface{}, rr *httptest.ResponseRecorder) {
	var body interface{}

	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Errorf("failed to parse body %v", err)
		return
	}

	expect.Equal(t, expected, body)
}
