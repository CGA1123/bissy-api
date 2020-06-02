package querycache_test

import (
	"database/sql"
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

// TODO
// func TestSQLAdapterUpdate(t *testing.T) {}
// func TestSQLAdapterList(t *testing.T) {}
