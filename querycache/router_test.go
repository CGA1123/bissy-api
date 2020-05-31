package querycache_test

import (
	"net/http"
	"testing"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/handlerutils"
)

func TestHome(t *testing.T) {
	t.Parallel()

	_, _, config := testConfig()
	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)

	expectedBody := "querycache: using cache, saving cash\n"

	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypePlaintext, response)
	expecthttp.StringBody(t, expectedBody, response)
}
