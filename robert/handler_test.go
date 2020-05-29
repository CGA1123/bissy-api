package robert

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/gorilla/mux"
)

func testHandler(c *Config, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	c.SetupHandlers(router)

	router.ServeHTTP(recorder, r)

	return recorder
}

func TestHome(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := "my-id"
	config := &Config{Store: newTestStore(now, id)}

	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expectedBody := "robert, a poor man's trevor\nrobert -> trebor -> trevor\n"

	expect.StatusOK(t, response)
	expect.ContentType(t, handlerutils.ContentTypePlaintext, response)
	expect.BodyString(t, expectedBody, response)
}
