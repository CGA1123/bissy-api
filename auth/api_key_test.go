package auth_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
	"github.com/jmoiron/sqlx"
)

func withTestSQLAPIKeyStore(t *testing.T, f func(*testing.T, auth.APIKeyStore, *auth.User, time.Time, string, string)) {
	db, err := sqlx.Open("pgx", uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	userID := uuid.New().String()

	userStore := testSQLUserStore(now, userID, hnysqlx.WrapDB(db))
	user, err := userStore.Create(&auth.CreateUser{GithubID: "github-id", Name: "test"})
	expect.Ok(t, err)

	key, err := (&utils.SecureRandom{}).String(32)
	expect.Ok(t, err)

	store := auth.NewTestSQLAPIKeyStore(hnysqlx.WrapDB(db), now, id, key)

	f(t, store, user, now, key, id)
}

func testAPIKeyCreate(t *testing.T, store auth.APIKeyStore, user *auth.User, now time.Time, key, id string) {
	expected := &auth.NewAPIKey{
		Name:      "test",
		UserID:    user.ID,
		ID:        id,
		Key:       key,
		CreatedAt: now,
		LastUsed:  now,
	}

	apiKey, err := store.Create(user.ID, &auth.CreateAPIKey{Name: "test"})
	expect.Ok(t, err)
	expect.Equal(t, expected, apiKey)

	// duplicate
	_, err = store.Create(user.ID, &auth.CreateAPIKey{Name: "test"})
	expect.Error(t, err)
}

func testAPIKeyDelete(t *testing.T, store auth.APIKeyStore, user *auth.User, now time.Time, key, id string) {
	apiKey, err := store.Create(user.ID, &auth.CreateAPIKey{Name: "test"})
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

func testAPIKeyGetByKey(t *testing.T, store auth.APIKeyStore, user *auth.User, now time.Time, key, id string) {
	apiKey, err := store.Create(user.ID, &auth.CreateAPIKey{Name: "test"})
	expect.Ok(t, err)

	expectedAPIKey := &auth.APIKey{
		ID:        apiKey.ID,
		UserID:    apiKey.UserID,
		Name:      apiKey.Name,
		LastUsed:  apiKey.LastUsed,
		CreatedAt: apiKey.CreatedAt}

	actual, err := store.GetByKey(apiKey.Key)
	expect.Ok(t, err)
	expect.Equal(t, expectedAPIKey, actual)

	// key doesn't exist
	_, err = store.GetByKey("nope")
	expect.Error(t, err)
	expect.True(t, err == sql.ErrNoRows)
}

func testAPIKeyList(t *testing.T, store auth.APIKeyStore, userOne, userTwo *auth.User) {
	expectedKeys := []*auth.APIKey{}
	for i := 0; i < 5; i++ {
		k := &auth.CreateAPIKey{Name: fmt.Sprintf("key %v", i)}
		apiKey, err := store.Create(userOne.ID, k)
		expect.Ok(t, err)

		expectedKeys = append(expectedKeys, &auth.APIKey{
			ID:        apiKey.ID,
			UserID:    apiKey.UserID,
			Name:      apiKey.Name,
			LastUsed:  apiKey.LastUsed,
			CreatedAt: apiKey.CreatedAt})
	}

	// with other user
	keys, err := store.List(userTwo.ID)
	expect.Ok(t, err)
	expect.Equal(t, []*auth.APIKey{}, keys)

	// with owning user
	keys, err = store.List(userOne.ID)
	expect.Ok(t, err)
	expect.Equal(t, expectedKeys, keys)

	// bad id
	_, err = store.List("")
	expect.Error(t, err)
}

func TestSQLAPIKeyStoreCreate(t *testing.T) {
	t.Parallel()

	withTestSQLAPIKeyStore(t, testAPIKeyCreate)
}

func TestSQLAPIKeyStoreDelete(t *testing.T) {
	t.Parallel()

	withTestSQLAPIKeyStore(t, testAPIKeyDelete)
}

func TestSQLAPIKeyStoreGetByKey(t *testing.T) {
	t.Parallel()

	withTestSQLAPIKeyStore(t, testAPIKeyGetByKey)
}

func TestSQLAPIKeyStoreList(t *testing.T) {
	t.Parallel()

	db, err := sqlx.Open("pgx", uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	hnydb := hnysqlx.WrapDB(db)
	userStore := auth.NewSQLUserStore(hnydb, &utils.RealClock{}, &utils.UUIDGenerator{})
	userOne, err := userStore.Create(&auth.CreateUser{GithubID: "github-id-1", Name: "test"})
	expect.Ok(t, err)

	userTwo, err := userStore.Create(&auth.CreateUser{GithubID: "github-id-2", Name: "test"})
	expect.Ok(t, err)

	store := auth.NewSQLAPIKeyStore(hnysqlx.WrapDB(db))
	expect.Ok(t, err)

	testAPIKeyList(t, store, userOne, userTwo)
}
