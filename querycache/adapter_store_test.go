package querycache_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func testAdapterCreate(t *testing.T, store querycache.AdapterStore, id string, now time.Time) {
	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test adapter",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	adapter, err := store.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable",
	})

	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)

	adapter, err = store.Get(id)
	expect.Ok(t, err)

	expect.Equal(t, expected, adapter)
}

func testAdapterDelete(t *testing.T, store querycache.AdapterStore, id string, now time.Time) {
	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test adapter",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := store.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	adapter, err := store.Delete(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)

	adapters, err := store.List(1, 1)
	expect.Ok(t, err)

	expect.Equal(t, []*querycache.Adapter{}, adapters)
}

func testAdapterGet(t *testing.T, store querycache.AdapterStore, id string, now time.Time) {
	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test adapter",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := store.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	adapter, err := store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)
}

func testAdapterUpdate(t *testing.T, store querycache.AdapterStore, id string, now time.Time) {
	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test snowdapter",
		Type:      "snowflake",
		Options:   "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := store.Create(&querycache.CreateAdapter{
		Name:    "test adapter",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	newName := "test snowdapter"
	newType := "snowflake"
	newOptions := ""
	adapter, err := store.Update(id, &querycache.UpdateAdapter{
		Name:    &newName,
		Type:    &newType,
		Options: &newOptions,
	})

	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)

	adapter, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)

	// partial update
	newName = "test snowdapter 2"
	expected.Name = newName
	adapter, err = store.Update(id, &querycache.UpdateAdapter{
		Name: &newName,
	})

	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)

	adapter, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)
}

func testAdapterList(t *testing.T, store querycache.AdapterStore) {
	expectedAdapters := []*querycache.Adapter{}

	for i := 0; i < 10; i++ {
		s := fmt.Sprintf("name %v;", i)
		q := &querycache.CreateAdapter{Name: s}
		adapter, err := store.Create(q)
		expect.Ok(t, err)

		expectedAdapters = append(expectedAdapters, adapter)
	}

	_, err := store.List(0, 1)
	expect.Error(t, err)

	_, err = store.List(1, 0)
	expect.Error(t, err)

	adapters, err := store.List(1, 10)
	expect.Ok(t, err)
	expect.Equal(t, expectedAdapters, adapters)

	adapters, err = store.List(2, 3)
	expect.Ok(t, err)
	expect.Equal(t, expectedAdapters[3:6], adapters)

	adapters, err = store.List(4, 3)
	expect.Ok(t, err)
	expect.Equal(t, expectedAdapters[9:10], adapters)

	adapters, err = store.List(10, 3)
	expect.Ok(t, err)
	expect.Equal(t, []*querycache.Adapter{}, adapters)

	adapters, err = store.List(1, 30)
	expect.Ok(t, err)
	expect.Equal(t, expectedAdapters, adapters)
}

func TestInMemoryAdapterCreate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestAdapterStore(now, id)

	testAdapterCreate(t, store, id, now)
}

func TestInMemoryAdapterUpdate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestAdapterStore(now, id)

	testAdapterUpdate(t, store, id, now)
}

func TestInMemoryAdapterDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestAdapterStore(now, id)

	testAdapterDelete(t, store, id, now)
}

func TestInMemoryAdapterList(t *testing.T) {
	t.Parallel()

	testAdapterList(t,
		querycache.NewInMemoryAdapterStore(&querycache.RealClock{}, &querycache.UUIDGenerator{}))
}

func TestInMemoryAdapterGet(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestAdapterStore(now, id)

	testAdapterGet(t, store, id, now)
}

func init() {
	url, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL not set")
	}

	txdb.Register("pgx", "postgres", url)
}

func testDb(id string) (*sqlx.DB, error) {
	return sqlx.Open("pgx", id)
}

func testSQLAdapterStore(now time.Time, id string, db *sqlx.DB) *querycache.SQLAdapterStore {
	return querycache.NewSQLAdapterStore(
		db,
		&testClock{time: now},
		&testIdGenerator{id: id},
	)
}

func withTestSQLAdapterStore(t *testing.T, f func(*testing.T, querycache.AdapterStore, string, time.Time)) {
	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	store := testSQLAdapterStore(now, id, db)

	f(t, store, id, now)
}

func TestSQLAdapterStoreCreate(t *testing.T) {
	t.Parallel()

	withTestSQLAdapterStore(t, testAdapterCreate)
}

func TestSQLAdapterGet(t *testing.T) {
	t.Parallel()

	withTestSQLAdapterStore(t, testAdapterGet)
}

func TestSQLAdapterDelete(t *testing.T) {
	t.Parallel()

	withTestSQLAdapterStore(t, testAdapterDelete)
}

func TestSQLAdapterList(t *testing.T) {
	t.Parallel()

	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	store := querycache.NewSQLAdapterStore(db,
		&querycache.RealClock{}, &querycache.UUIDGenerator{})

	testAdapterList(t, store)
}

func TestSQLAdapterUpdate(t *testing.T) {
	t.Parallel()

	withTestSQLAdapterStore(t, testAdapterUpdate)
}
