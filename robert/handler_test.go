package robert

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// TODO: move to expect.BodyStructJson() ?
// maybe should have expecthttp tbh separate core and http helpers
func structAsJson(data interface{}) (interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var mapData interface{}
	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil, err
	}

	return mapData, nil
}

func testHandler(c *Config, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	c.SetupHandlers(router)

	router.ServeHTTP(recorder, r)

	return recorder
}

func testConfig() (time.Time, string, *Config) {
	now := time.Now()
	id := uuid.New().String()

	return now, id, &Config{Store: newTestStore(now, id)}
}

func jsonBody(v interface{}) (*bytes.Reader, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(body), nil
}

func TestHome(t *testing.T) {
	t.Parallel()

	_, _, config := testConfig()
	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)

	expectedBody := "robert, a poor man's trevor\nrobert -> trebor -> trevor\n"

	expect.StatusOK(t, response)
	expect.ContentType(t, handlerutils.ContentTypePlaintext, response)
	expect.BodyString(t, expectedBody, response)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	now, id, config := testConfig()
	json, err := jsonBody(map[string]string{
		"lifetime": "1h01m",
		"query":    "SELECT 1;",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("POST", "/queries", json)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expect.StatusOK(t, response)

	actual, err := config.Store.Get(id)
	expected := &Query{
		Id:          id,
		Lifetime:    Duration(time.Hour + time.Minute),
		Query:       "SELECT 1;",
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	expect.Ok(t, err)
	expect.Equal(t, expected, actual)
}

func TestCreateBadRequest(t *testing.T) {
	t.Parallel()

	body := strings.NewReader("not json")
	request, err := http.NewRequest("POST", "/queries", body)
	expect.Ok(t, err)

	_, _, config := testConfig()
	response := testHandler(config, request)

	expect.StatusHTTP(t, http.StatusUnprocessableEntity, response)
}

// TODO: Handle not found errors, add custom errors to store and switch?
// can add to store -> handler error function as well to help!
func TestGet(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	query, err := config.Store.Create(&CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: Duration(time.Hour),
	})
	expect.Ok(t, err)

	queryAsJson, err := structAsJson(query)
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/queries/"+id, nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expect.StatusOK(t, response)
	expect.ContentType(t, handlerutils.ContentTypeJson, response)
	expect.BodyJSON(t, queryAsJson, response)
}
