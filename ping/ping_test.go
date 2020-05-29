package ping

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/handlerutils"
)

func TestHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health-check", nil)
	expect.Ok(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)

	handler.ServeHTTP(rr, req)

	expectedJson := map[string]interface{}{"message": "pong"}

	expect.StatusOK(t, rr)
	expect.ContentType(t, handlerutils.ContentTypeJson, rr)
	expect.BodyJSON(t, expectedJson, rr)
}
