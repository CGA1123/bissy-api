package slackerduty_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cga1123/bissy-api/slackerduty"
	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/expect"
	"github.com/cga1123/bissy-api/utils/expecthttp"
	"github.com/gorilla/mux"
)

func testHandler(c *slackerduty.Config, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	c.SetupHandlers(router)

	router.ServeHTTP(recorder, r)

	return recorder
}

func TestPagerdutyEvent(t *testing.T) {
	t.Parallel()

	config := &slackerduty.Config{
		PagerdutyWebhookToken: "test-token"}

	request, err := http.NewRequest("POST", "/pagerduty/event/not-test-token", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Status(t, http.StatusUnauthorized, response)

	request, err = http.NewRequest("POST", "/pagerduty/event/test-token", nil)

	response = testHandler(config, request)
	expecthttp.Status(t, http.StatusBadRequest, response)

	json, err := utils.JSONBody(map[string]interface{}{
		"messages": []interface{}{}})
	request, err = http.NewRequest("POST", "/pagerduty/event/test-token", json)
	response = testHandler(config, request)
	expecthttp.Ok(t, response)
}
