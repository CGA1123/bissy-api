package apikeyprovider_test

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/auth/apikey"
	"github.com/cga1123/bissy-api/auth/apikeyprovider"
	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/expect"
	_ "github.com/lib/pq"
)

func init() {
	url, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL not set")
	}

	txdb.Register("pgx", "postgres", url)
}

func TestIsProvider(t *testing.T) {
	t.Parallel()

	db, teardown := utils.TestDB(t)
	defer teardown()

	store := apikey.NewSQLStore(db)
	config := apikeyprovider.New(store)

	// check we conform to Provider interface
	// tests will fail to compile if not.
	var _ auth.Provider = config
}

func TestValid(t *testing.T) {
	t.Parallel()

	db, teardown := utils.TestDB(t)
	defer teardown()

	store := apikey.NewSQLStore(db)
	config := apikeyprovider.New(store)

	// empty request
	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	expect.False(t, config.Valid(request))

	// request with header
	request.Header.Add(apikeyprovider.HeaderKey, "token")

	expect.True(t, config.Valid(request))
}

func unsuccessfulAuth(t *testing.T, c *apikeyprovider.Config, r *http.Request) {
	t.Helper()

	claims, ok := c.Authenticate(r)

	expect.True(t, nil == claims)
	expect.False(t, ok)
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	db, teardown := utils.TestDB(t)
	defer teardown()

	user, err := auth.NewSQLUserStore(db).Create(&auth.CreateUser{GithubID: "test", Name: "Test"})
	expect.Ok(t, err)

	store := apikey.NewSQLStore(db)
	key, err := store.Create(user.ID, &apikey.Create{Name: "test key"})
	expect.Ok(t, err)

	config := apikeyprovider.New(store)

	// empty request
	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	unsuccessfulAuth(t, config, request)

	// unknown key
	request, err = http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	request.Header.Add(apikeyprovider.HeaderKey, "an unknown key")

	unsuccessfulAuth(t, config, request)

	// known key
	request, err = http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	request.Header.Add(apikeyprovider.HeaderKey, key.Key)

	claims, ok := config.Authenticate(request)
	expect.True(t, ok)
	expect.Equal(t, user.ID, claims.UserID)
}
