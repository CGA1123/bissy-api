package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

func testingHandler(t *testing.T, expected *auth.User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(*auth.User)

		expect.True(t, ok)
		expect.Equal(t, expected, user)
	})
}

func testHandler(c *auth.Config, r *http.Request, h http.Handler) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.
		Handle("/", h).
		Methods("GET")

	c.WithAuth(router).ServeHTTP(recorder, r)

	return recorder
}

func TestAuthHandler(t *testing.T) {
	t.Parallel()

	db, err := sqlx.Open("pgx", uuid.New().String())
	expect.Ok(t, err)
	defer db.Close()

	err = db.Ping()
	expect.Ok(t, err)

	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()
	store := testSQLUserStore(now, id, db)
	user, err := store.Create(&auth.CreateUser{Email: "test@bissy.io"})
	expect.Ok(t, err)

	signingKey := []byte("test-key")
	clock := &testClock{time: now}
	config := auth.NewConfig(signingKey, time.Hour, store, clock)

	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	r := testHandler(config, request, testingHandler(t, user))
	expecthttp.Status(t, http.StatusUnauthorized, r)
	expecthttp.Header(t, "WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`, r)

	// with bad auth header
	request, err = http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)
	request.Header.Add("Authorization", "Bearer 123453")

	r = testHandler(config, request, testingHandler(t, user))
	expecthttp.Status(t, http.StatusUnauthorized, r)
	expecthttp.Header(t, "WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`, r)

	// with correct auth header
	token, err := config.SignedToken(user)
	expect.Ok(t, err)

	request, err = http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)
	request.Header.Add("Authorization", "Bearer "+token)
	r = testHandler(config, request, testingHandler(t, user))
	expecthttp.Ok(t, r)
	expecthttp.Header(t, "Bissy-Token", token, r)
}
