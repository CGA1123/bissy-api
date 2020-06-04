package auth_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/expect"
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

type testClock struct {
	time time.Time
}

func (c *testClock) Now() time.Time {
	return c.time
}

type testIdGenerator struct {
	id string
}

func (g *testIdGenerator) Generate() string {
	return g.id
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

	// fail if email already exists
	_, err = store.Create(&auth.CreateUser{Email: "test@bissy.io"})
	expect.Error(t, err)
}

func testSQLUserStore(now time.Time, id string, db *sqlx.DB) *auth.SQLUserStore {
	return auth.NewSQLUserStore(
		db,
		&testClock{time: now},
		&testIdGenerator{id: id},
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
