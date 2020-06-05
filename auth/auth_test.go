package auth_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func init() {
	url, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL not set")
	}

	txdb.Register("pgx", "postgres", url)
}

func testUserGet(t *testing.T, store auth.UserStore, id string, now time.Time) {
	expected, err := store.Create(&auth.CreateUser{Email: "test@bissy.io"})
	expect.Ok(t, err)

	user, err := store.Get(expected.Id)
	expect.Ok(t, err)
	expect.Equal(t, expected, user)

	_, err = store.Get(uuid.New().String())
	expect.Error(t, err)
}

func testUserCreate(t *testing.T, store auth.UserStore, id string, now time.Time) {
	user, err := store.Create(&auth.CreateUser{Email: "test@bissy.io"})
	expect.Ok(t, err)

	expected := &auth.User{Id: id, Email: "test@bissy.io", CreatedAt: now}
	expect.Equal(t, expected, user)

	user, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, user)
}

func testUserEmailDuplicate(t *testing.T, store auth.UserStore) {
	_, err := store.Create(&auth.CreateUser{Email: "test@bissy.io"})
	expect.Ok(t, err)

	_, err = store.Create(&auth.CreateUser{Email: "test@bissy.io"})
	expect.Error(t, err)
}

func testSQLUserStore(now time.Time, id string, db *sqlx.DB) *auth.SQLUserStore {
	return auth.NewSQLUserStore(
		db,
		&utils.TestClock{Time: now},
		&utils.TestIdGenerator{Id: id},
	)
}

func withTestSQLUserStore(t *testing.T, f func(*testing.T, auth.UserStore, string, time.Time)) {
	db, err := sqlx.Open("pgx", uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	store := testSQLUserStore(now, id, db)

	f(t, store, id, now)
}

func TestSQLUserGet(t *testing.T) {
	t.Parallel()

	withTestSQLUserStore(t, testUserGet)
}

func TestSQLUserCreate(t *testing.T) {
	t.Parallel()

	withTestSQLUserStore(t, testUserCreate)
}

func TestSQLUserEmailDuplicate(t *testing.T) {
	db, err := sqlx.Open("pgx", uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	store := auth.NewSQLUserStore(db, &utils.RealClock{}, &utils.UUIDGenerator{})

	testUserEmailDuplicate(t, store)
}
