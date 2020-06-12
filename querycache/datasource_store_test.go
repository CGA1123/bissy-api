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
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
	"github.com/jmoiron/sqlx"
)

func testDatasourceCreate(t *testing.T, store querycache.DatasourceStore, id string, now time.Time) {
	expected := &querycache.Datasource{
		Id:        id,
		Name:      "test datasource",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	datasource, err := store.Create(&querycache.CreateDatasource{
		Name:    "test datasource",
		Type:    "postgres",
		Options: "sslmode=disable",
	})

	expect.Ok(t, err)
	expect.Equal(t, expected, datasource)

	datasource, err = store.Get(id)
	expect.Ok(t, err)

	expect.Equal(t, expected, datasource)
}

func testDatasourceDelete(t *testing.T, store querycache.DatasourceStore, id string, now time.Time) {
	expected := &querycache.Datasource{
		Id:        id,
		Name:      "test datasource",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := store.Create(&querycache.CreateDatasource{
		Name:    "test datasource",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	datasource, err := store.Delete(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, datasource)

	datasources, err := store.List(1, 1)
	expect.Ok(t, err)

	expect.Equal(t, []*querycache.Datasource{}, datasources)
}

func testDatasourceGet(t *testing.T, store querycache.DatasourceStore, id string, now time.Time) {
	expected := &querycache.Datasource{
		Id:        id,
		Name:      "test datasource",
		Type:      "postgres",
		Options:   "sslmode=disable",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := store.Create(&querycache.CreateDatasource{
		Name:    "test datasource",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	datasource, err := store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, datasource)
}

func testDatasourceUpdate(t *testing.T, store querycache.DatasourceStore, id string, now time.Time) {
	expected := &querycache.Datasource{
		Id:        id,
		Name:      "test snowdapter",
		Type:      "snowflake",
		Options:   "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := store.Create(&querycache.CreateDatasource{
		Name:    "test datasource",
		Type:    "postgres",
		Options: "sslmode=disable",
	})
	expect.Ok(t, err)

	newName := "test snowdapter"
	newType := "snowflake"
	newOptions := ""
	datasource, err := store.Update(id, &querycache.UpdateDatasource{
		Name:    &newName,
		Type:    &newType,
		Options: &newOptions,
	})

	expect.Ok(t, err)
	expect.Equal(t, expected, datasource)

	datasource, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, datasource)

	// partial update
	newName = "test snowdapter 2"
	expected.Name = newName
	datasource, err = store.Update(id, &querycache.UpdateDatasource{
		Name: &newName,
	})

	expect.Ok(t, err)
	expect.Equal(t, expected, datasource)

	datasource, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, datasource)
}

func testDatasourceList(t *testing.T, store querycache.DatasourceStore) {
	expectedDatasources := []*querycache.Datasource{}

	for i := 0; i < 10; i++ {
		s := fmt.Sprintf("name %v;", i)
		q := &querycache.CreateDatasource{Name: s}
		datasource, err := store.Create(q)
		expect.Ok(t, err)

		expectedDatasources = append(expectedDatasources, datasource)
	}

	_, err := store.List(0, 1)
	expect.Error(t, err)

	_, err = store.List(1, 0)
	expect.Error(t, err)

	datasources, err := store.List(1, 10)
	expect.Ok(t, err)
	expect.Equal(t, expectedDatasources, datasources)

	datasources, err = store.List(2, 3)
	expect.Ok(t, err)
	expect.Equal(t, expectedDatasources[3:6], datasources)

	datasources, err = store.List(4, 3)
	expect.Ok(t, err)
	expect.Equal(t, expectedDatasources[9:10], datasources)

	datasources, err = store.List(10, 3)
	expect.Ok(t, err)
	expect.Equal(t, []*querycache.Datasource{}, datasources)

	datasources, err = store.List(1, 30)
	expect.Ok(t, err)
	expect.Equal(t, expectedDatasources, datasources)
}

func TestInMemoryDatasourceCreate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestDatasourceStore(now, id)

	testDatasourceCreate(t, store, id, now)
}

func TestInMemoryDatasourceUpdate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestDatasourceStore(now, id)

	testDatasourceUpdate(t, store, id, now)
}

func TestInMemoryDatasourceDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestDatasourceStore(now, id)

	testDatasourceDelete(t, store, id, now)
}

func TestInMemoryDatasourceList(t *testing.T) {
	t.Parallel()

	testDatasourceList(t,
		querycache.NewInMemoryDatasourceStore(&utils.RealClock{}, &utils.UUIDGenerator{}))
}

func TestInMemoryDatasourceGet(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestDatasourceStore(now, id)

	testDatasourceGet(t, store, id, now)
}

func init() {
	url, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL not set")
	}

	txdb.Register("pgx", "postgres", url)
}

func testDb(id string) (*hnysqlx.DB, error) {
	db, err := sqlx.Open("pgx", id)
	return hnysqlx.WrapDB(db), err
}

func testSQLDatasourceStore(now time.Time, id string, db *hnysqlx.DB) *querycache.SQLDatasourceStore {
	return querycache.NewSQLDatasourceStore(
		db,
		&utils.TestClock{Time: now},
		&utils.TestIdGenerator{Id: id},
	)
}

func withTestSQLDatasourceStore(t *testing.T, f func(*testing.T, querycache.DatasourceStore, string, time.Time)) {
	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	store := testSQLDatasourceStore(now, id, db)

	f(t, store, id, now)
}

func TestSQLDatasourceStoreCreate(t *testing.T) {
	t.Parallel()

	withTestSQLDatasourceStore(t, testDatasourceCreate)
}

func TestSQLDatasourceGet(t *testing.T) {
	t.Parallel()

	withTestSQLDatasourceStore(t, testDatasourceGet)
}

func TestSQLDatasourceDelete(t *testing.T) {
	t.Parallel()

	withTestSQLDatasourceStore(t, testDatasourceDelete)
}

func TestSQLDatasourceList(t *testing.T) {
	t.Parallel()

	db, err := testDb(uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	store := querycache.NewSQLDatasourceStore(db,
		&utils.RealClock{}, &utils.UUIDGenerator{})

	testDatasourceList(t, store)
}

func TestSQLDatasourceUpdate(t *testing.T) {
	t.Parallel()

	withTestSQLDatasourceStore(t, testDatasourceUpdate)
}