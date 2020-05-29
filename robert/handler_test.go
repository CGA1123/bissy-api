package robert

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	return now, id, &Config{
		Store:    newTestStore(now, id),
		Executor: &testExecutor{}}
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

// TODO: Handle not found errors, add custom errors to store and switch?
// can add to store -> handler error function as well to help!
func TestGetNotFound(t *testing.T) {
	t.Parallel()

	_, _, config := testConfig()

	request, err := http.NewRequest("GET", "/queries/does-not-exist", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expect.StatusHTTP(t, http.StatusInternalServerError, response)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	query, err := config.Store.Create(&CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: Duration(time.Hour),
	})
	expect.Ok(t, err)

	queryAsJson, err := structAsJson(query)
	expect.Ok(t, err)

	request, err := http.NewRequest("DELETE", "/queries/"+id, nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expect.StatusOK(t, response)
	expect.ContentType(t, handlerutils.ContentTypeJson, response)
	expect.BodyJSON(t, queryAsJson, response)

	queries, err := config.Store.List(1, 1)
	expect.Ok(t, err)
	expect.Equal(t, []Query{}, queries)
}

// TODO: Handle not found errors, add custom errors to store and switch?
// can add to store -> handler error function as well to help!
func TestDeleteNotFound(t *testing.T) {
	t.Parallel()

	_, _, config := testConfig()

	request, err := http.NewRequest("DELETE", "/queries/does-not-exist", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expect.StatusHTTP(t, http.StatusInternalServerError, response)
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	query, err := config.Store.Create(&CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: Duration(time.Hour),
	})
	expect.Ok(t, err)

	json, err := jsonBody(map[string]string{
		"lifetime": "1h01m",
		"query":    "SELECT 2;",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("PATCH", "/queries/"+id, json)
	expect.Ok(t, err)

	query.Lifetime = Duration(time.Hour + time.Minute)
	query.Query = "SELECT 2;"
	queryAsJson, err := structAsJson(query)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expect.StatusOK(t, response)
	expect.ContentType(t, handlerutils.ContentTypeJson, response)
	expect.BodyJSON(t, queryAsJson, response)

	query, err = config.Store.Get(id)
	expect.Ok(t, err)

	expect.Equal(t, Duration(time.Hour+time.Minute), query.Lifetime)
	expect.Equal(t, "SELECT 2;", query.Query)
}

// TODO: move to expect.BodyStructJson() ?
// maybe should have expecthttp tbh separate core and http helpers
func TestUpdateNotFound(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	json, err := jsonBody(map[string]string{
		"lifetime": "1h01m",
		"query":    "SELECT 2;",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("PATCH", "/queries/"+id, json)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expect.StatusHTTP(t, http.StatusInternalServerError, response)
}

func TestList(t *testing.T) {
	t.Parallel()

	config := &Config{Store: NewInMemoryStore(&RealClock{}, &UUIDGenerator{})}
	queries := []Query{}

	for i := 0; i < 30; i++ {
		query, err := config.Store.Create(&CreateQuery{
			Query: fmt.Sprintf("SELECT %v", i)})

		expect.Ok(t, err)
		queries = append(queries, *query)
	}

	request, err := http.NewRequest("GET", "/queries", nil)
	expect.Ok(t, err)

	expected, err := structAsJson(queries[:25])
	expect.Ok(t, err)

	response := testHandler(config, request)
	expect.StatusOK(t, response)
	expect.ContentType(t, handlerutils.ContentTypeJson, response)
	expect.BodyJSON(t, expected, response)

	// pagination
	request, err = http.NewRequest("GET", "/queries?page=2&per=5", nil)
	expect.Ok(t, err)

	expected, err = structAsJson(queries[5:10])
	expect.Ok(t, err)

	response = testHandler(config, request)
	expect.StatusOK(t, response)
	expect.ContentType(t, handlerutils.ContentTypeJson, response)
	expect.BodyJSON(t, expected, response)
}

func TestExecute(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()

	_, err := config.Store.Create(&CreateQuery{Query: "SELECT * FROM users"})
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/queries/"+id+"/result", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expect.StatusOK(t, response)
	expect.ContentType(t, handlerutils.ContentTypeCsv, response)
	expect.BodyString(t, "Got: SELECT * FROM users", response)
}
