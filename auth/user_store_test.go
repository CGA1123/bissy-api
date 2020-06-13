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
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
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

func testUserGetByGithubID(t *testing.T, store auth.UserStore, id string, now time.Time) {
	expected, err := store.Create(&auth.CreateUser{GithubID: "github-id", Name: "Test"})
	expect.Ok(t, err)

	user, err := store.GetByGithubID(expected.GithubID)
	expect.Ok(t, err)
	expect.Equal(t, expected, user)

	_, err = store.GetByGithubID(uuid.New().String())
	expect.Error(t, err)
}

func testUserGet(t *testing.T, store auth.UserStore, id string, now time.Time) {
	expected, err := store.Create(&auth.CreateUser{GithubID: "github-id", Name: "Test"})
	expect.Ok(t, err)

	user, err := store.Get(expected.ID)
	expect.Ok(t, err)
	expect.Equal(t, expected, user)

	_, err = store.Get(uuid.New().String())
	expect.Error(t, err)
}

func testUserCreate(t *testing.T, store auth.UserStore, id string, now time.Time) {
	user, err := store.Create(&auth.CreateUser{GithubID: "github-id", Name: "Test"})
	expect.Ok(t, err)

	expected := &auth.User{ID: id, GithubID: "github-id", Name: "Test", CreatedAt: now}
	expect.Equal(t, expected, user)

	user, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, user)
}

func testUserEmailDuplicate(t *testing.T, store auth.UserStore) {
	_, err := store.Create(&auth.CreateUser{GithubID: "github-id", Name: "Test"})
	expect.Ok(t, err)

	_, err = store.Create(&auth.CreateUser{GithubID: "github-id", Name: "Test"})
	expect.Error(t, err)
}

func testSQLUserStore(now time.Time, id string, db *hnysqlx.DB) *auth.SQLUserStore {
	return auth.NewSQLUserStore(
		db,
		&utils.TestClock{Time: now},
		&utils.TestIDGenerator{ID: id},
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
	store := testSQLUserStore(now, id, hnysqlx.WrapDB(db))

	f(t, store, id, now)
}

func TestSQLUserGet(t *testing.T) {
	t.Parallel()

	withTestSQLUserStore(t, testUserGet)
}

func TestSQLUserGetByGithubID(t *testing.T) {
	t.Parallel()

	withTestSQLUserStore(t, testUserGetByGithubID)
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

	store := auth.NewSQLUserStore(hnysqlx.WrapDB(db), &utils.RealClock{}, &utils.UUIDGenerator{})

	testUserEmailDuplicate(t, store)
}
