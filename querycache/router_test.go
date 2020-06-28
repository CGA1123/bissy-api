package querycache_test

import (
	"net/http"
	"testing"

	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/expect"
	"github.com/cga1123/bissy-api/utils/expecthttp"
	"github.com/cga1123/bissy-api/utils/handlerutils"
)

func TestHome(t *testing.T) {
	t.Parallel()

	db, teardown := utils.TestDB(t)
	defer teardown()

	_, _, config := testConfig(db)
	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	claims := testClaims()
	response := testHandler(claims, config, request)

	expectedBody := "querycache: using cache, saving cash\n"

	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypePlaintext, response)
	expecthttp.StringBody(t, expectedBody, response)
}
