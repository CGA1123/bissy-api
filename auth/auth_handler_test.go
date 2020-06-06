package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func testConfig(t *testing.T, now time.Time, userId, redisId, clientId string) (*auth.Config, *auth.SQLUserStore, *auth.RedisStore, func()) {
	db, dbTeardown := utils.TestDB(t)
	redisClient, redisTeardown := utils.TestRedis(t)
	redis := &auth.RedisStore{Client: redisClient}

	store := testSQLUserStore(now.Truncate(time.Millisecond), userId, db)
	signingKey := []byte("test-key")
	clock := &utils.TestClock{Time: now}
	config := auth.NewConfig(
		signingKey,
		time.Hour,
		store,
		clock,
		redis,
		&utils.TestIdGenerator{Id: redisId},
		clientId,
	)

	return config, store, redis, func() {
		expect.Ok(t, dbTeardown())
		expect.Ok(t, redisTeardown())
	}
}

func testingHandler(t *testing.T, expected *auth.User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(*auth.User)

		expect.True(t, ok)
		expect.Equal(t, expected, user)
	})
}

func testRouter(c *auth.Config, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	c.SetupHandlers(router)

	router.ServeHTTP(recorder, r)

	return recorder
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

	now := time.Now().Truncate(time.Millisecond)
	userId := uuid.New().String()
	config, store, _, teardown := testConfig(t, now, userId, uuid.New().String(), uuid.New().String())
	defer teardown()

	user, err := store.Create(&auth.CreateUser{Email: "test@bissy.io"})
	expect.Ok(t, err)

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

func TestGithubSignIn(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	userId := uuid.New().String()
	clientId := uuid.New().String()
	redisId := uuid.New().String()
	config, _, redis, teardown := testConfig(t, now, userId, redisId, clientId)
	defer teardown()

	request, err := http.NewRequest("GET", "/github/signin", nil)
	expect.Ok(t, err)

	response := testRouter(config, request)

	githubRedirectUrl := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%v&state=%v&scope=user",
		clientId, redisId)
	expecthttp.Status(t, http.StatusTemporaryRedirect, response)
	expecthttp.Header(t, "Location", githubRedirectUrl, response)

	exists, err := redis.Exists(redisId)
	expect.Ok(t, err)
	expect.True(t, exists)
}

func TestGithubCallback(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	userId := uuid.New().String()
	clientId := uuid.New().String()
	redisId := uuid.New().String()
	config, _, redis, teardown := testConfig(t, now, userId, redisId, clientId)
	defer teardown()

	err := redis.Set(redisId, time.Hour)
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/github/callback?code=my-code&state="+redisId, nil)
	expect.Ok(t, err)

	response := testRouter(config, request)
	expecthttp.Ok(t, response)
}
