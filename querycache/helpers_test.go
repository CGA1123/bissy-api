package querycache_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cga1123/bissy-api/querycache"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type testClock struct {
	time time.Time
}

func (clock *testClock) Now() time.Time {
	return clock.time
}

type testIdGenerator struct {
	id string
}

func (generator *testIdGenerator) Generate() string {
	return generator.id
}

func newTestQueryStore(now time.Time, id string) *querycache.InMemoryQueryStore {
	return querycache.NewInMemoryQueryStore(
		&testClock{time: now},
		&testIdGenerator{id: id})
}

func newTestAdapterStore(now time.Time, id string) *querycache.InMemoryAdapterStore {
	return querycache.NewInMemoryAdapterStore(
		&testClock{time: now},
		&testIdGenerator{id: id})
}

func testCachedExecutor() *querycache.CachedExecutor {
	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)
	return &querycache.CachedExecutor{
		Cache:    querycache.NewInMemoryCache(),
		Store:    store,
		Executor: &querycache.TestExecutor{},
		Clock:    &testClock{time: now},
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
