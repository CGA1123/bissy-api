package querycache_test

import (
	"database/sql"
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

func TestSQLAdapterStoreCreate(t *testing.T) {
	t.Parallel()

	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	store := testSQLAdapterStore(now, id, db)
	adapter, err := store.Create(&querycache.CreateAdapter{
		Name:    "Test",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	expected := &querycache.Adapter{
		Id:        id,
		Name:      "Test",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	expect.Equal(t, expected, adapter)
}

func TestSQLAdapterGet(t *testing.T) {
	t.Parallel()

	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	store := testSQLAdapterStore(now, id, db)
	expected, err := store.Create(&querycache.CreateAdapter{
		Name:    "Test",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	adapter, err := store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)
}

func TestSQLAdapterDelete(t *testing.T) {
	t.Parallel()

	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	store := testSQLAdapterStore(now, id, db)
	expected, err := store.Create(&querycache.CreateAdapter{
		Name:    "Test",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	adapter, err := store.Delete(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, adapter)

	_, err = store.Get(id)
	expect.True(t, err == sql.ErrNoRows)
}

func TestSQLAdapterList(t *testing.T) {
	t.Parallel()

	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	store := querycache.NewSQLAdapterStore(db,
		&querycache.RealClock{}, &querycache.UUIDGenerator{})

	expectedAdapters := []*querycache.Adapter{}

	for i := 0; i < 10; i++ {
		s := fmt.Sprintf("name %v;", i)
		q := &querycache.CreateAdapter{Name: s}
		adapter, err := store.Create(q)
		expect.Ok(t, err)

		expectedAdapters = append(expectedAdapters, adapter)
	}

	_, err = store.List(0, 1)
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

func TestSQLAdapterUpdate(t *testing.T) {
	t.Parallel()

	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	store := testSQLAdapterStore(now, id, db)
	expected := &querycache.Adapter{
		Id:        id,
		Name:      "test snowdapter",
		Type:      "snowflake",
		Options:   "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = store.Create(&querycache.CreateAdapter{
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
	newName = "test snowdapter-2"
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
