package querycache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
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

func testQueryCreate(t *testing.T, datasourceStore querycache.DatasourceStore, store querycache.QueryStore, id string, now time.Time) {
	datasource, err := datasourceStore.Create(&querycache.CreateDatasource{})
	expect.Ok(t, err)

	createQuery := querycache.CreateQuery{
		Query:        "SELECT 1;",
		Lifetime:     3 * querycache.Duration(time.Hour),
		DatasourceId: datasource.Id,
	}

	expected := &querycache.Query{
		Id:           id,
		Query:        "SELECT 1;",
		DatasourceId: datasource.Id,
		Lifetime:     3 * querycache.Duration(time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
		LastRefresh:  now,
	}

	query, err := store.Create(&createQuery)

	expect.Ok(t, err)
	expect.Equal(t, expected, query)

	query, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, query)
}

func testQueryGet(t *testing.T, datasourceStore querycache.DatasourceStore, store querycache.QueryStore, id string, now time.Time) {
	datasource, err := datasourceStore.Create(&querycache.CreateDatasource{})
	expect.Ok(t, err)

	createQuery := querycache.CreateQuery{
		Query:        "SELECT 1;",
		Lifetime:     3 * querycache.Duration(time.Hour),
		DatasourceId: datasource.Id,
	}

	expected := querycache.Query{
		Id:           id,
		Query:        "SELECT 1;",
		DatasourceId: datasource.Id,
		Lifetime:     3 * querycache.Duration(time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
		LastRefresh:  now,
	}

	_, err = store.Create(&createQuery)
	expect.Ok(t, err)

	// when id is found
	query, err := store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, *query)

	// when id is not found
	_, err = store.Get(uuid.New().String())
	expect.Error(t, err)
}

func testQueryList(t *testing.T, datasourceStore querycache.DatasourceStore, store querycache.QueryStore) {
	datasource, err := datasourceStore.Create(&querycache.CreateDatasource{})
	expect.Ok(t, err)

	expectedQueries := []*querycache.Query{}

	for i := 0; i < 10; i++ {
		q := &querycache.CreateQuery{
			Query:        fmt.Sprintf("SELECT %v;", i),
			Lifetime:     querycache.Duration(time.Duration(i) * time.Hour),
			DatasourceId: datasource.Id,
		}
		query, err := store.Create(q)
		expect.Ok(t, err)

		expectedQueries = append(expectedQueries, query)
	}

	_, err = store.List(0, 1)
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

func testQueryDelete(t *testing.T, datasourceStore querycache.DatasourceStore, store querycache.QueryStore, id string, now time.Time) {
	datasource, err := datasourceStore.Create(&querycache.CreateDatasource{})
	expect.Ok(t, err)

	createQuery := querycache.CreateQuery{
		Query:        "SELECT 1;",
		Lifetime:     3 * querycache.Duration(time.Hour),
		DatasourceId: datasource.Id,
	}

	expected := querycache.Query{
		Id:           id,
		Query:        "SELECT 1;",
		DatasourceId: datasource.Id,
		Lifetime:     3 * querycache.Duration(time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
		LastRefresh:  now,
	}

	_, err = store.Create(&createQuery)
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

func testQueryUpdate(t *testing.T, datasourceStore querycache.DatasourceStore, store querycache.QueryStore, id string, now time.Time) {
	datasource, err := datasourceStore.Create(&querycache.CreateDatasource{})
	expect.Ok(t, err)

	createQuery := querycache.CreateQuery{
		Query:        "SELECT 1;",
		DatasourceId: datasource.Id,
		Lifetime:     3 * querycache.Duration(time.Hour),
	}

	datasourceTwo, err := datasourceStore.Create(&querycache.CreateDatasource{})
	expect.Ok(t, err)

	newQuery := "SELECT 2;"
	newLifetime := querycache.Duration(time.Hour)
	newDatasourceId := datasourceTwo.Id
	updateQuery := querycache.UpdateQuery{
		Query:        &newQuery,
		Lifetime:     &newLifetime,
		DatasourceId: &newDatasourceId,
	}

	expected := querycache.Query{
		Id:           id,
		Query:        "SELECT 2;",
		DatasourceId: datasourceTwo.Id,
		Lifetime:     querycache.Duration(time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
		LastRefresh:  now,
	}

	_, err = store.Create(&createQuery)
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
	_, err = store.Update(uuid.New().String(), &updateQuery)
	expect.Error(t, err)
}

func TestInMemoryQueryCreate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)
	datasourceStore := querycache.NewInMemoryDatasourceStore(&utils.RealClock{}, &utils.UUIDGenerator{})

	testQueryCreate(t, datasourceStore, store, id, now)
}

func TestInMemoryQueryGet(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)
	datasourceStore := querycache.NewInMemoryDatasourceStore(&utils.RealClock{}, &utils.UUIDGenerator{})

	testQueryGet(t, datasourceStore, store, id, now)
}

func TestInMemoryQueryDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)
	datasourceStore := querycache.NewInMemoryDatasourceStore(&utils.RealClock{}, &utils.UUIDGenerator{})

	testQueryDelete(t, datasourceStore, store, id, now)
}

func TestInMemoryQueryUpdate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestQueryStore(now, id)
	datasourceStore := querycache.NewInMemoryDatasourceStore(&utils.RealClock{}, &utils.UUIDGenerator{})

	testQueryUpdate(t, datasourceStore, store, id, now)
}

func TestInMemoryQueryList(t *testing.T) {
	t.Parallel()

	store := querycache.NewInMemoryQueryStore(&utils.RealClock{}, &utils.UUIDGenerator{})
	datasourceStore := querycache.NewInMemoryDatasourceStore(&utils.RealClock{}, &utils.UUIDGenerator{})

	testQueryList(t, datasourceStore, store)
}

func testSQLQueryStore(now time.Time, id string, db *hnysqlx.DB) *querycache.SQLQueryStore {
	return querycache.NewSQLQueryStore(
		db,
		&utils.TestClock{Time: now},
		&utils.TestIdGenerator{Id: id},
	)
}

func withTestSQLQueryStore(t *testing.T, f func(*testing.T, querycache.DatasourceStore, querycache.QueryStore, string, time.Time)) {
	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	store := testSQLQueryStore(now, id, db)
	datasourceStore := querycache.NewSQLDatasourceStore(db, &utils.RealClock{}, &utils.UUIDGenerator{})

	f(t, datasourceStore, store, id, now)
}

func TestSQLQueryCreate(t *testing.T) {
	t.Parallel()

	withTestSQLQueryStore(t, testQueryCreate)
}

func TestSQLQueryGet(t *testing.T) {
	t.Parallel()

	withTestSQLQueryStore(t, testQueryGet)
}

func TestSQLQueryDelete(t *testing.T) {
	t.Parallel()

	withTestSQLQueryStore(t, testQueryDelete)
}

func TestSQLQueryUpdate(t *testing.T) {
	t.Parallel()

	withTestSQLQueryStore(t, testQueryUpdate)
}

func TestSQLQueryList(t *testing.T) {
	t.Parallel()

	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)
	store := querycache.NewSQLQueryStore(db, &utils.RealClock{}, &utils.UUIDGenerator{})
	datasourceStore := querycache.NewSQLDatasourceStore(db, &utils.RealClock{}, &utils.UUIDGenerator{})

	testQueryList(t, datasourceStore, store)
}
