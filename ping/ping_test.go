package ping

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/handlerutils"
)

func TestHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health-check", nil)
	expect.Ok(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)

	handler.ServeHTTP(rr, req)

	expectedJSON := map[string]interface{}{"message": "pong"}

	expecthttp.Ok(t, rr)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, rr)
	expecthttp.JSONBody(t, expectedJSON, rr.Body)
}
