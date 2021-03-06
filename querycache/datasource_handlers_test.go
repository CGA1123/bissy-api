package querycache_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cga1123/bissy-api/querycache"
	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/expect"
	"github.com/cga1123/bissy-api/utils/expecthttp"
	"github.com/cga1123/bissy-api/utils/handlerutils"
)

func TestDatasourceCreate(t *testing.T) {
	t.Parallel()

	db, teardown := utils.TestDB(t)
	defer teardown()

	now, id, config := testConfig(db)
	json, err := utils.JSONBody(map[string]string{
		"name":    "test datasource",
		"type":    "postgres",
		"options": "",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("POST", "/datasources", json)
	expect.Ok(t, err)

	claims := testClaims()
	expected := &querycache.Datasource{
		ID:        id,
		UserID:    claims.UserID,
		Name:      "test datasource",
		Type:      "postgres",
		Options:   "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	response := testHandler(claims, config, request)
	expecthttp.Ok(t, response)
	expecthttp.JSONBody(t, expected, response.Body)

	actual, err := config.DatasourceStore.Get(claims.UserID, id)

	expect.Ok(t, err)
	expect.Equal(t, expected, actual)
}

func TestDatasourceGet(t *testing.T) {
	t.Parallel()

	db, teardown := utils.TestDB(t)
	defer teardown()

	_, id, config := testConfig(db)
	claims := testClaims()
	datasource, err := config.DatasourceStore.Create(claims.UserID, &querycache.CreateDatasource{
		Name:    "test datasource",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/datasources/"+id, nil)
	expect.Ok(t, err)

	response := testHandler(claims, config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, datasource, response.Body)
}

func TestDatasourceDelete(t *testing.T) {
	t.Parallel()

	db, teardown := utils.TestDB(t)
	defer teardown()

	_, id, config := testConfig(db)
	claims := testClaims()
	datasource, err := config.DatasourceStore.Create(claims.UserID, &querycache.CreateDatasource{
		Name:    "test datasource",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("DELETE", "/datasources/"+id, nil)
	expect.Ok(t, err)

	response := testHandler(claims, config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, datasource, response.Body)

	datasources, err := config.DatasourceStore.List(claims.UserID, 1, 1)
	expect.Ok(t, err)
	expect.Equal(t, []*querycache.Datasource{}, datasources)
}

func TestDatasourceUpdate(t *testing.T) {
	t.Parallel()

	db, teardown := utils.TestDB(t)
	defer teardown()

	_, id, config := testConfig(db)
	claims := testClaims()
	datasource, err := config.DatasourceStore.Create(claims.UserID, &querycache.CreateDatasource{
		Name:    "test datasource",
		Type:    "postgres",
		Options: "sslmode=disable"})
	expect.Ok(t, err)

	json, err := utils.JSONBody(map[string]string{
		"name":    "test",
		"type":    "snowflake",
		"options": ""})
	expect.Ok(t, err)

	request, err := http.NewRequest("PATCH", "/datasources/"+id, json)
	expect.Ok(t, err)

	datasource.Name = "test"
	datasource.Type = "snowflake"
	datasource.Options = ""

	response := testHandler(claims, config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, datasource, response.Body)

	datasource, err = config.DatasourceStore.Get(claims.UserID, id)
	expect.Ok(t, err)

	expect.Equal(t, "test", datasource.Name)
	expect.Equal(t, "snowflake", datasource.Type)
	expect.Equal(t, "", datasource.Options)
}

func TestDatasourceList(t *testing.T) {
	t.Parallel()

	datasources := []*querycache.Datasource{}

	db, teardown := utils.TestDB(t)
	defer teardown()

	claims := testClaims()
	config := &querycache.Config{
		DatasourceStore: querycache.NewSQLDatasourceStore(db, &utils.RealClock{}, &utils.UUIDGenerator{})}

	for i := 0; i < 30; i++ {
		datasource, err := config.DatasourceStore.Create(claims.UserID, &querycache.CreateDatasource{
			Name: fmt.Sprintf("Name %v", i)})

		expect.Ok(t, err)
		datasources = append(datasources, datasource)
	}

	request, err := http.NewRequest("GET", "/datasources", nil)
	expect.Ok(t, err)

	response := testHandler(claims, config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, datasources[:25], response.Body)

	// pagination
	request, err = http.NewRequest("GET", "/datasources?page=2&per=5", nil)
	expect.Ok(t, err)

	response = testHandler(claims, config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
	expecthttp.JSONBody(t, datasources[5:10], response.Body)
}
