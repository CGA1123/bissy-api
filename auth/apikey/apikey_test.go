package apikey_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/auth/apikey"
	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/expect"
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

func withTestSQLStore(t *testing.T, f func(*testing.T, apikey.Store, *auth.User, time.Time, string, string)) {
	db, err := sqlx.Open("pgx", uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	userID := uuid.New().String()

	userStore := auth.TestSQLUserStore(now, userID, hnysqlx.WrapDB(db))
	user, err := userStore.Create(&auth.CreateUser{GithubID: "github-id", Name: "test"})
	expect.Ok(t, err)

	key, err := (&utils.SecureRandom{}).String(32)
	expect.Ok(t, err)

	store := apikey.NewTestSQLStore(hnysqlx.WrapDB(db), now, id, key)

	f(t, store, user, now, key, id)
}

func testCreate(t *testing.T, store apikey.Store, user *auth.User, now time.Time, key, id string) {
	expected := &apikey.New{
		Name:      "test",
		UserID:    user.ID,
		ID:        id,
		Key:       key,
		CreatedAt: now,
		LastUsed:  now,
	}

	apiKey, err := store.Create(user.ID, &apikey.Create{Name: "test"})
	expect.Ok(t, err)
	expect.Equal(t, expected, apiKey)

	// duplicate
	_, err = store.Create(user.ID, &apikey.Create{Name: "test"})
	expect.Error(t, err)
}

func testDelete(t *testing.T, store apikey.Store, user *auth.User, now time.Time, key, id string) {
	apiKey, err := store.Create(user.ID, &apikey.Create{Name: "test"})
	expect.Ok(t, err)

	// different user
	_, err = store.Delete(uuid.New().String(), apiKey.ID)
	expect.Error(t, err)
	expect.True(t, err == sql.ErrNoRows)

	_, err = store.GetByKey(apiKey.Key)
	expect.Ok(t, err)

	// key doesn't exist
	_, err = store.Delete(user.ID, uuid.New().String())
	expect.Error(t, err)
	expect.True(t, err == sql.ErrNoRows)

	_, err = store.GetByKey(apiKey.Key)
	expect.Ok(t, err)

	// key exists
	deletedKey, err := store.Delete(user.ID, apiKey.ID)
	expect.Ok(t, err)
	expect.Equal(t, apiKey.ID, deletedKey.ID)

	_, err = store.GetByKey(apiKey.Key)
	expect.Error(t, err)
	expect.True(t, err == sql.ErrNoRows)
}

func testGetByKey(t *testing.T, store apikey.Store, user *auth.User, now time.Time, key, id string) {
	apiKey, err := store.Create(user.ID, &apikey.Create{Name: "test"})
	expect.Ok(t, err)

	expected := &apikey.Struct{
		ID:        apiKey.ID,
		UserID:    apiKey.UserID,
		Name:      apiKey.Name,
		LastUsed:  apiKey.LastUsed,
		CreatedAt: apiKey.CreatedAt}

	actual, err := store.GetByKey(apiKey.Key)
	expect.Ok(t, err)
	expect.Equal(t, expected, actual)

	// key doesn't exist
	_, err = store.GetByKey("nope")
	expect.Error(t, err)
	expect.True(t, err == sql.ErrNoRows)
}

func testList(t *testing.T, store apikey.Store, userOne, userTwo *auth.User) {
	expectedKeys := []*apikey.Struct{}
	for i := 0; i < 5; i++ {
		k := &apikey.Create{Name: fmt.Sprintf("key %v", i)}
		apiKey, err := store.Create(userOne.ID, k)
		expect.Ok(t, err)

		expectedKeys = append(expectedKeys, &apikey.Struct{
			ID:        apiKey.ID,
			UserID:    apiKey.UserID,
			Name:      apiKey.Name,
			LastUsed:  apiKey.LastUsed,
			CreatedAt: apiKey.CreatedAt})
	}

	// with other user
	keys, err := store.List(userTwo.ID)
	expect.Ok(t, err)
	expect.Equal(t, []*apikey.Struct{}, keys)

	// with owning user
	keys, err = store.List(userOne.ID)
	expect.Ok(t, err)
	expect.Equal(t, expectedKeys, keys)

	// bad id
	_, err = store.List("")
	expect.Error(t, err)
}

func TestSQLStoreCreate(t *testing.T) {
	t.Parallel()

	withTestSQLStore(t, testCreate)
}

func TestSQLStoreDelete(t *testing.T) {
	t.Parallel()

	withTestSQLStore(t, testDelete)
}

func TestSQLStoreGetByKey(t *testing.T) {
	t.Parallel()

	withTestSQLStore(t, testGetByKey)
}

func TestSQLStoreList(t *testing.T) {
	t.Parallel()

	db, err := sqlx.Open("pgx", uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	hnydb := hnysqlx.WrapDB(db)
	userStore := auth.NewSQLUserStore(hnydb)
	userOne, err := userStore.Create(&auth.CreateUser{GithubID: "github-id-1", Name: "test"})
	expect.Ok(t, err)

	userTwo, err := userStore.Create(&auth.CreateUser{GithubID: "github-id-2", Name: "test"})
	expect.Ok(t, err)

	store := apikey.NewSQLStore(hnysqlx.WrapDB(db))
	expect.Ok(t, err)

	testList(t, store, userOne, userTwo)
}
