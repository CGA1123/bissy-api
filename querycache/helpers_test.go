package querycache_test

import (
	"fmt"
	"time"

	"github.com/cga1123/bissy-api/querycache"
	"github.com/google/uuid"
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

func newTestStore(now time.Time, id string) *querycache.InMemoryStore {
	return querycache.NewInMemoryStore(
		&testClock{time: now},
		&testIdGenerator{id: id})
}

type testExecutor struct{}

func (t *testExecutor) Execute(query *querycache.Query) (string, error) {
	return fmt.Sprintf("Got: %v", query.Query), nil
}

func testCachedExecutor() *querycache.CachedExecutor {
	now := time.Now()
	id := uuid.New().String()
	store := newTestStore(now, id)
	return &querycache.CachedExecutor{
		Cache:    querycache.NewInMemoryCache(),
		Store:    store,
		Executor: &testExecutor{},
		Clock:    &testClock{time: now},
	}
}
