package querycache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/google/uuid"
)

func freshQuery(now time.Time) *querycache.Query {
	oneHourAgo := now.Add(-time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	return &querycache.Query{
		LastRefresh: oneHourAgo,
		UpdatedAt:   twoHoursAgo,
		Lifetime:    querycache.Duration(3 * time.Hour)}
}

func TestFresh(t *testing.T) {
	t.Parallel()

	now := time.Now()
	var query *querycache.Query

	// fresh query
	query = freshQuery(now)
	expect.True(t, query.Fresh(now))

	// query updated after last refresh
	query = freshQuery(now)
	query.UpdatedAt = query.LastRefresh.Add(time.Second)
	expect.False(t, query.Fresh(now))

	// last refresh longer than lifetime ago
	query = freshQuery(now)
	query.LastRefresh = now.Add(-time.Duration(query.Lifetime)).Add(-time.Second)
	expect.False(t, query.Fresh(now))

	// last refresh longer than lifetime ago and updated after last refresh
	query = freshQuery(now)
	query.UpdatedAt = query.LastRefresh.Add(time.Second)
	query.LastRefresh = now.Add(-time.Duration(query.Lifetime)).Add(-time.Second)
	expect.False(t, query.Fresh(now))
}

func testQueryCreate(t *testing.T, store querycache.QueryStore, id string, now time.Time) {
	createQuery := querycache.CreateQuery{
		Query:     "SELECT 1;",
		Lifetime:  3 * querycache.Duration(time.Hour),
		AdapterId: "my-adapter",
	}

	expected := &querycache.Query{
		Id:          id,
		Query:       "SELECT 1;",
		AdapterId:   "my-adapter",
		Lifetime:    3 * querycache.Duration(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	query, err := store.Create(&createQuery)

	expect.Ok(t, err)
	expect.Equal(t, expected, query)

	query, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, query)
}

func testQueryGet(t *testing.T, store querycache.QueryStore, id string, now time.Time) {
	createQuery := querycache.CreateQuery{
		Query:     "SELECT 1;",
		Lifetime:  3 * querycache.Duration(time.Hour),
		AdapterId: "adapter-id",
	}

	expected := querycache.Query{
		Id:          id,
		Query:       "SELECT 1;",
		AdapterId:   "adapter-id",
		Lifetime:    3 * querycache.Duration(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	_, err := store.Create(&createQuery)
	expect.Ok(t, err)

	// when id is found
	query, err := store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, *query)

	// when id is not found
	_, err = store.Get("")
	expect.Error(t, err)
}

func testQueryList(t *testing.T, store querycache.QueryStore) {
	expectedQueries := []*querycache.Query{}

	for i := 0; i < 10; i++ {
		q := &querycache.CreateQuery{
			Query:    fmt.Sprintf("SELECT %v;", i),
			Lifetime: querycache.Duration(time.Duration(i) * time.Hour),
		}
		query, err := store.Create(q)
		expect.Ok(t, err)

		expectedQueries = append(expectedQueries, query)
	}

	_, err := store.List(0, 1)
	expect.Error(t, err)

	_, err = store.List(1, 0)
	expect.Error(t, err)

	queries, err := store.List(1, 10)
	expect.Ok(t, err)
	expect.Equal(t, expectedQueries, queries)

	queries, err = store.List(2, 3)
	expect.Ok(t, err)
	expect.Equal(t, expectedQueries[3:6], queries)

	queries, err = store.List(4, 3)
	expect.Ok(t, err)
	expect.Equal(t, expectedQueries[9:10], queries)

	queries, err = store.List(10, 3)
	expect.Ok(t, err)
	expect.Equal(t, []*querycache.Query{}, queries)

	queries, err = store.List(1, 30)
	expect.Ok(t, err)
	expect.Equal(t, expectedQueries, queries)
}

func testQueryDelete(t *testing.T, store querycache.QueryStore, id string, now time.Time) {
	createQuery := querycache.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: 3 * querycache.Duration(time.Hour),
	}

	expected := querycache.Query{
		Id:          id,
		Query:       "SELECT 1;",
		Lifetime:    3 * querycache.Duration(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	_, err := store.Create(&createQuery)
	expect.Ok(t, err)

	query, err := store.Delete(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, *query)

	queries, err := store.List(1, 1)
	expect.Ok(t, err)

	expect.Equal(t, []*querycache.Query{}, queries)

	_, err = store.Delete(id)
	expect.Error(t, err)
}

func testQueryUpdate(t *testing.T, store querycache.QueryStore, id string, now time.Time) {
	createQuery := querycache.CreateQuery{
		Query:     "SELECT 1;",
		AdapterId: "adapter-1",
		Lifetime:  3 * querycache.Duration(time.Hour),
	}

	newQuery := "SELECT 2;"
	newLifetime := querycache.Duration(time.Hour)
	newAdapterId := "adapter-2"
	updateQuery := querycache.UpdateQuery{
		Query:     &newQuery,
		Lifetime:  &newLifetime,
		AdapterId: &newAdapterId,
	}

	expected := querycache.Query{
		Id:          id,
		Query:       "SELECT 2;",
		AdapterId:   "adapter-2",
		Lifetime:    querycache.Duration(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	_, err := store.Create(&createQuery)
	expect.Ok(t, err)

	// Test returned query
	query, err := store.Update(id, &updateQuery)
	expect.Ok(t, err)
	expect.Equal(t, expected, *query)

	// Test persistence
	query, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, *query)

	// Test partial update
	newQuery = "SELECT 15;"
	query, err = store.Update(id, &querycache.UpdateQuery{Query: &newQuery})
	expect.Ok(t, err)
	expect.Equal(t, newQuery, query.Query)
	expect.Equal(t, newLifetime, query.Lifetime)

	newLifetime = 15 * querycache.Duration(time.Second)
	query, err = store.Update(id, &querycache.UpdateQuery{Lifetime: &newLifetime})
	expect.Ok(t, err)
	expect.Equal(t, newLifetime, query.Lifetime)
	expect.Equal(t, newQuery, query.Query)

	// Test updating lastrefresh
	query, err = store.Update(id, &querycache.UpdateQuery{LastRefresh: now.Add(time.Hour)})
	expect.Ok(t, err)
	expect.Equal(t, newQuery, query.Query)
	expect.Equal(t, newLifetime, query.Lifetime)
	expect.Equal(t, now.Add(time.Hour), query.LastRefresh)

	// Updating not existing query
	_, err = store.Update("", &updateQuery)
	expect.Error(t, err)
}

func TestInMemoryQueryCreate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)

	testQueryCreate(t, store, id, now)
}

func TestInMemoryQueryGet(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)

	testQueryGet(t, store, id, now)
}

func TestInMemoryQueryDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)

	testQueryDelete(t, store, id, now)
}

func TestInMemoryQueryUpdate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)

	testQueryUpdate(t, store, id, now)
}

func TestInMemoryQueryList(t *testing.T) {
	t.Parallel()

	store := querycache.NewInMemoryQueryStore(&querycache.RealClock{}, &querycache.UUIDGenerator{})

	testQueryList(t, store)
}
