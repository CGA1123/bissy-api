package querycache_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/querycache"
	_ "github.com/lib/pq"
)

func TestQueryCreate(t *testing.T) {
	t.Parallel()

	now, id, config := testConfig()
	json, err := jsonBody(map[string]string{
		"lifetime":  "1h01m",
		"query":     "SELECT 1;",
		"adapterId": "adapter-id",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("POST", "/queries", json)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)

	actual, err := config.QueryStore.Get(id)
	expected := &querycache.Query{
		Id:          id,
		Lifetime:    querycache.Duration(time.Hour + time.Minute),
		Query:       "SELECT 1;",
		AdapterId:   "adapter-id",
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	expect.Ok(t, err)
	expect.Equal(t, expected, actual)
}

func TestQueryCreateBadRequest(t *testing.T) {
	t.Parallel()

	body := strings.NewReader("not json")
	request, err := http.NewRequest("POST", "/queries", body)
	expect.Ok(t, err)

	_, _, config := testConfig()
	response := testHandler(config, request)

	expecthttp.Status(t, http.StatusUnprocessableEntity, response)
}

func TestQueryGet(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	query, err := config.QueryStore.Create(&querycache.CreateQuery{
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
func TestQueryGetNotFound(t *testing.T) {
	t.Parallel()

	_, _, config := testConfig()

	request, err := http.NewRequest("GET", "/queries/does-not-exist", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Status(t, http.StatusInternalServerError, response)
}

func TestQueryDelete(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	query, err := config.QueryStore.Create(&querycache.CreateQuery{
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

	queries, err := config.QueryStore.List(1, 1)
	expect.Ok(t, err)
	expect.Equal(t, []querycache.Query{}, queries)
}

// TODO: Handle not found errors, add custom errors to store and switch?
// can add to store -> handler error function as well to help!
func TestQueryDeleteNotFound(t *testing.T) {
	t.Parallel()

	_, _, config := testConfig()

	request, err := http.NewRequest("DELETE", "/queries/does-not-exist", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Status(t, http.StatusInternalServerError, response)
}

func TestQueryUpdate(t *testing.T) {
	t.Parallel()

	now, id, config := testConfig()
	query, err := config.QueryStore.Create(&querycache.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: querycache.Duration(time.Hour),
	})
	expect.Ok(t, err)

	oneHourAgo := now.Add(-time.Hour)
	json, err := jsonBody(map[string]string{
		"lifetime":    "1h01m",
		"query":       "SELECT 2;",
		"adapterId":   "adapter-id",
		"lastRefresh": oneHourAgo.Format(time.RFC3339Nano)})
	expect.Ok(t, err)

	request, err := http.NewRequest("PATCH", "/queries/"+id, json)
	expect.Ok(t, err)

	query.Lifetime = querycache.Duration(time.Hour + time.Minute)
	query.Query = "SELECT 2;"
	query.LastRefresh = oneHourAgo
	query.AdapterId = "adapter-id"

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, query, response)

	query, err = config.QueryStore.Get(id)
	expect.Ok(t, err)

	expect.Equal(t, querycache.Duration(time.Hour+time.Minute), query.Lifetime)
	expect.Equal(t, "SELECT 2;", query.Query)
}

func TestQueryUpdateNotFound(t *testing.T) {
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

func TestQueryList(t *testing.T) {
	t.Parallel()

	queries := []querycache.Query{}
	config := &querycache.Config{
		QueryStore: querycache.NewInMemoryQueryStore(&querycache.RealClock{},
			&querycache.UUIDGenerator{})}

	for i := 0; i < 30; i++ {
		query, err := config.QueryStore.Create(&querycache.CreateQuery{
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

func TestQueryExecute(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	_, err := config.AdapterStore.Create(&querycache.CreateAdapter{Type: "test", Name: "Test"})
	expect.Ok(t, err)

	_, err = config.QueryStore.Create(&querycache.CreateQuery{Query: "SELECT * FROM users", AdapterId: id})
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/queries/"+id+"/result", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeCsv, response)
	expecthttp.StringBody(t, "Got: SELECT * FROM users", response)

	// PG Test
	newType := "postgres"
	newName := "PG Test"
	newOptions := "sslmode=disable"
	_, err = config.AdapterStore.Update(id, &querycache.UpdateAdapter{
		Type: &newType, Name: &newName, Options: &newOptions})

	expect.Ok(t, err)

	newQuery := "SELECT 1"
	_, err = config.QueryStore.Update(id, &querycache.UpdateQuery{
		Query: &newQuery})
	expect.Ok(t, err)

	request, err = http.NewRequest("GET", "/queries/"+id+"/result", nil)
	expect.Ok(t, err)

	response = testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeCsv, response)
	expecthttp.StringBody(t, "?column?\n1\n", response)
}
