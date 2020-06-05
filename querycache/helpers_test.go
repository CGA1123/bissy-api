package querycache_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cga1123/bissy-api/querycache"
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func newTestQueryStore(now time.Time, id string) *querycache.InMemoryQueryStore {
	return querycache.NewInMemoryQueryStore(
		&utils.TestClock{Time: now},
		&utils.TestIdGenerator{Id: id})
}

func newTestAdapterStore(now time.Time, id string) *querycache.InMemoryAdapterStore {
	return querycache.NewInMemoryAdapterStore(
		&utils.TestClock{Time: now},
		&utils.TestIdGenerator{Id: id})
}

func testCachedExecutor() *querycache.CachedExecutor {
	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)
	return &querycache.CachedExecutor{
		Cache:    querycache.NewInMemoryCache(),
		Store:    store,
		Executor: &querycache.TestExecutor{},
		Clock:    &utils.TestClock{Time: now},
	}
}

func testConfig() (time.Time, string, *querycache.Config) {
	now := time.Now()
	id := uuid.New().String()

	return now, id, &querycache.Config{
		QueryStore:   newTestQueryStore(now, id),
		AdapterStore: newTestAdapterStore(now, id),
		Executor:     &querycache.TestExecutor{}}
}

func testHandler(c *querycache.Config, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	c.SetupHandlers(router)

	router.ServeHTTP(recorder, r)

	return recorder
}

func jsonBody(v interface{}) (*bytes.Reader, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(body), nil
}
