package querycache_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/cga1123/bissy-api/utils"
	_ "github.com/lib/pq"
)

func TestQueryCreate(t *testing.T) {
	t.Parallel()

	now, id, config := testConfig()
	json, err := jsonBody(map[string]string{
		"lifetime":     "1h01m",
		"query":        "SELECT 1;",
		"datasourceID": "datasource-id",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("POST", "/queries", json)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)

	actual, err := config.QueryStore.Get(id)
	expected := &querycache.Query{
		ID:           id,
		Lifetime:     querycache.Duration(time.Hour + time.Minute),
		Query:        "SELECT 1;",
		DatasourceID: "datasource-id",
		CreatedAt:    now,
		UpdatedAt:    now,
		LastRefresh:  now,
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
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, query, response.Body)
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
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, query, response.Body)

	queries, err := config.QueryStore.List(1, 1)
	expect.Ok(t, err)
	expect.Equal(t, []*querycache.Query{}, queries)
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
		"lifetime":     "1h01m",
		"query":        "SELECT 2;",
		"datasourceID": "datasource-id",
		"lastRefresh":  oneHourAgo.Format(time.RFC3339Nano)})
	expect.Ok(t, err)

	request, err := http.NewRequest("PATCH", "/queries/"+id, json)
	expect.Ok(t, err)

	query.Lifetime = querycache.Duration(time.Hour + time.Minute)
	query.Query = "SELECT 2;"
	query.LastRefresh = oneHourAgo
	query.DatasourceID = "datasource-id"

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, query, response.Body)

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
		QueryStore: querycache.NewInMemoryQueryStore(&utils.RealClock{},
			&utils.UUIDGenerator{})}

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
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, queries[:25], response.Body)

	// pagination
	request, err = http.NewRequest("GET", "/queries?page=2&per=5", nil)
	expect.Ok(t, err)

	response = testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, queries[5:10], response.Body)
}

func TestQueryResult(t *testing.T) {
	t.Parallel()

	clock := &utils.RealClock{}
	generator := &utils.UUIDGenerator{}
	config := &querycache.Config{
		QueryStore:      querycache.NewInMemoryQueryStore(clock, generator),
		DatasourceStore: querycache.NewInMemoryDatasourceStore(clock, generator),
	}

	datasource, err := config.DatasourceStore.Create(&querycache.CreateDatasource{Type: "test", Name: "Test"})
	expect.Ok(t, err)

	query, err := config.QueryStore.Create(&querycache.CreateQuery{
		Query: "SELECT * FROM users", DatasourceID: datasource.ID})
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/queries/"+query.ID+"/result", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeCSV, response)
	expecthttp.StringBody(t, "Got: SELECT * FROM users", response)

	// PG Test
	newType := "postgres"
	newName := "PG Test"
	newOptions, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		t.Fatal("DATABASE_URL is not set")
	}

	datasource, err = config.DatasourceStore.Update(datasource.ID, &querycache.UpdateDatasource{
		Type: &newType, Name: &newName, Options: &newOptions})

	expect.Ok(t, err)

	newQuery := "SELECT 1"
	query, err = config.QueryStore.Update(query.ID, &querycache.UpdateQuery{
		Query: &newQuery})
	expect.Ok(t, err)

	request, err = http.NewRequest("GET", "/queries/"+query.ID+"/result", nil)
	expect.Ok(t, err)

	response = testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeCSV, response)
	expecthttp.StringBody(t, "?column?\n1\n", response)
}
