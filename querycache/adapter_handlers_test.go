package querycache_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/cga1123/bissy-api/utils"
)

func TestAdapterCreate(t *testing.T) {
	t.Parallel()

	now, id, config := testConfig()
	json, err := jsonBody(map[string]string{
		"name":    "test adapter",
		"type":    "postgres",
		"options": "",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("POST", "/adapters", json)
	expect.Ok(t, err)

	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test adapter",
		Type:      "postgres",
		Options:   "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.JSONBody(t, expected, response.Body)

	actual, err := config.AdapterStore.Get(id)

	expect.Ok(t, err)
	expect.Equal(t, expected, actual)
}

func TestAdapterGet(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	adapter, err := config.AdapterStore.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/adapters/"+id, nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, adapter, response.Body)
}

func TestAdapterDelete(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	adapter, err := config.AdapterStore.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	request, err := http.NewRequest("DELETE", "/adapters/"+id, nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, adapter, response.Body)

	adapters, err := config.AdapterStore.List(1, 1)
	expect.Ok(t, err)
	expect.Equal(t, []*querycache.Adapter{}, adapters)
}

func TestAdapterUpdate(t *testing.T) {
	t.Parallel()

	_, id, config := testConfig()
	adapter, err := config.AdapterStore.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable"})
	expect.Ok(t, err)

	json, err := jsonBody(map[string]string{
		"name":    "test",
		"type":    "snowflake",
		"options": ""})
	expect.Ok(t, err)

	request, err := http.NewRequest("PATCH", "/adapters/"+id, json)
	expect.Ok(t, err)

	adapter.Name = "test"
	adapter.Type = "snowflake"
	adapter.Options = ""

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, adapter, response.Body)

	adapter, err = config.AdapterStore.Get(id)
	expect.Ok(t, err)

	expect.Equal(t, "test", adapter.Name)
	expect.Equal(t, "snowflake", adapter.Type)
	expect.Equal(t, "", adapter.Options)
}

func TestAdapterList(t *testing.T) {
	t.Parallel()

	adapters := []*querycache.Adapter{}
	config := &querycache.Config{
		AdapterStore: querycache.NewInMemoryAdapterStore(&utils.RealClock{},
			&utils.UUIDGenerator{})}

	for i := 0; i < 30; i++ {
		adapter, err := config.AdapterStore.Create(&querycache.CreateAdapter{
			Name: fmt.Sprintf("Name %v", i)})

		expect.Ok(t, err)
		adapters = append(adapters, adapter)
	}

	request, err := http.NewRequest("GET", "/adapters", nil)
	expect.Ok(t, err)

	response := testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, adapters[:25], response.Body)

	// pagination
	request, err = http.NewRequest("GET", "/adapters?page=2&per=5", nil)
	expect.Ok(t, err)

	response = testHandler(config, request)
	expecthttp.Ok(t, response)
	expecthttp.ContentType(t, handlerutils.ContentTypeJson, response)
	expecthttp.JSONBody(t, adapters[5:10], response.Body)
}
