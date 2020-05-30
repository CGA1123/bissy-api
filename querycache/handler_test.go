package querycache_test

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
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func testHandler(c *querycache.Config, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	c.SetupHandlers(router)

	router.ServeHTTP(recorder, r)

	return recorder
}

func testConfig() (time.Time, string, *querycache.Config) {
	now := time.Now()
	id := uuid.New().String()

	return now, id, &querycache.Config{
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

	expectedBody := "querycache, a poor man's trevor\nquerycache -> trebor -> trevor\n"

	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypePlaintext, response)
	expecthttp.StringBody(t, expectedBody, response)
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
	expecthttp.Ok(t, response)

	actual, err := config.Store.Get(id)
	expected := &querycache.Query{
		Id:          id,
		Lifetime:    querycache.Duration(time.Hour + time.Minute),
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

	expecthttp.Status(t, http.StatusUnprocessableEntity, response)
}

func TestGet(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	query, err := config.Store.Create(&querycache.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: querycache.Duration(time.Hour),
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/queries/"+id, nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, query, response)
}

// TODO: Handle not found errors, add custom errors to store and switch?
// can add to store -> handler error function as well to help!
func TestGetNotFound(t *testing.T) {
	t.Parallel()

	_, _, config := testConfig()

	request, err := http.NewRequest("GET", "/queries/does-not-exist", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Status(t, http.StatusInternalServerError, response)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	query, err := config.Store.Create(&querycache.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: querycache.Duration(time.Hour),
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("DELETE", "/queries/"+id, nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, query, response)

	queries, err := config.Store.List(1, 1)
	expect.Ok(t, err)
	expect.Equal(t, []querycache.Query{}, queries)
}

// TODO: Handle not found errors, add custom errors to store and switch?
// can add to store -> handler error function as well to help!
func TestDeleteNotFound(t *testing.T) {
	t.Parallel()

	_, _, config := testConfig()

	request, err := http.NewRequest("DELETE", "/queries/does-not-exist", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Status(t, http.StatusInternalServerError, response)
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	now, id, config := testConfig()
	query, err := config.Store.Create(&querycache.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: querycache.Duration(time.Hour),
	})
	expect.Ok(t, err)

	oneHourAgo := now.Add(-time.Hour)
	json, err := jsonBody(map[string]string{
		"lifetime":    "1h01m",
		"query":       "SELECT 2;",
		"lastRefresh": oneHourAgo.Format(time.RFC3339Nano)})
	expect.Ok(t, err)

	request, err := http.NewRequest("PATCH", "/queries/"+id, json)
	expect.Ok(t, err)

	query.Lifetime = querycache.Duration(time.Hour + time.Minute)
	query.Query = "SELECT 2;"
	query.LastRefresh = oneHourAgo

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, query, response)

	query, err = config.Store.Get(id)
	expect.Ok(t, err)

	expect.Equal(t, querycache.Duration(time.Hour+time.Minute), query.Lifetime)
	expect.Equal(t, "SELECT 2;", query.Query)
}

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
	expecthttp.Status(t, http.StatusInternalServerError, response)
}

func TestList(t *testing.T) {
	t.Parallel()

	config := &querycache.Config{Store: querycache.NewInMemoryStore(&querycache.RealClock{}, &querycache.UUIDGenerator{})}
	queries := []querycache.Query{}

	for i := 0; i < 30; i++ {
		query, err := config.Store.Create(&querycache.CreateQuery{
			Query: fmt.Sprintf("SELECT %v", i)})

		expect.Ok(t, err)
		queries = append(queries, *query)
	}

	request, err := http.NewRequest("GET", "/queries", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, queries[:25], response)

	// pagination
	request, err = http.NewRequest("GET", "/queries?page=2&per=5", nil)
	expect.Ok(t, err)

	response = testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, queries[5:10], response)
}

func TestExecute(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()

	_, err := config.Store.Create(&querycache.CreateQuery{Query: "SELECT * FROM users"})
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/queries/"+id+"/result", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeCsv, response)
	expecthttp.StringBody(t, "Got: SELECT * FROM users", response)
}
