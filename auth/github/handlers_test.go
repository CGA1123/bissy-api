package github_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/auth/github"
	"github.com/cga1123/bissy-api/auth/jwtprovider"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func init() {
	url, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL not set")
	}

	txdb.Register("pgx", "postgres", url)
}

func testConfig(t *testing.T, now time.Time, userID, redisID string, client utils.HTTPClient) (*github.Config, *jwtprovider.Config, *auth.SQLUserStore, *github.RedisStateStore, func()) {
	db, dbTeardown := utils.TestDB(t)
	redisClient, redisTeardown := utils.TestRedis(t)
	redis := &github.RedisStateStore{Client: redisClient, IDGenerator: &utils.TestIDGenerator{ID: redisID}}
	githubApp := github.NewApp("client-id", "client-secret", client)

	store := auth.TestSQLUserStore(now.Truncate(time.Millisecond), userID, db)
	signingKey := []byte("test-key")
	authConfig := jwtprovider.TestConfig(signingKey, now)
	config := github.TestConfig(
		authConfig,
		store,
		redis,
		githubApp,
		now,
	)

	return config, authConfig, store, redis, func() {
		expect.Ok(t, dbTeardown())
		expect.Ok(t, redisTeardown())
	}
}

func testingHandler(t *testing.T, expected *auth.User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claim, ok := auth.UserFromContext(r.Context())

		expect.True(t, ok)
		expect.Equal(t, expected.ID, claim.UserID)
		expect.Equal(t, expected.Name, claim.Name)
	})
}

func testRouter(c *github.Config, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	c.SetupHandlers(router)

	router.ServeHTTP(recorder, r)

	return recorder
}

func testHandler(c *auth.Auth, r *http.Request, h http.Handler) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.
		Handle("/", h).
		Methods("GET")

	c.Middleware(router).ServeHTTP(recorder, r)

	return recorder
}

func TestTokenHandler(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	userID := uuid.New().String()
	redisID := uuid.New().String()
	httpClient := utils.NewTestHTTPClient()
	config, authConfig, store, redis, teardown := testConfig(t, now, userID, redisID, httpClient)
	defer teardown()

	user, err := store.Create(&auth.CreateUser{GithubID: "github-id", Name: "Chritian"})
	expect.Ok(t, err)

	redisKey, err := redis.Set(user.ID, time.Minute*5)
	expect.Ok(t, err)

	token, err := authConfig.SignedToken(user)
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/token?code="+redisKey, nil)
	expect.Ok(t, err)

	response := testRouter(config, request)
	expecthttp.Ok(t, response)
	expecthttp.JSONBody(t, map[string]interface{}{"token": token}, response.Body)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)

	// code expired
	_, err = redis.Del(redisKey)
	expect.Ok(t, err)

	request, err = http.NewRequest("GET", "/token?code="+redisKey, nil)
	expect.Ok(t, err)

	response = testRouter(config, request)
	expecthttp.Status(t, http.StatusBadRequest, response)
	expecthttp.StringBody(t, "bad token\n", response)

	// code linked to non-existing user
	redisKey, err = redis.Set(uuid.New().String(), time.Minute*5)
	expect.Ok(t, err)

	request, err = http.NewRequest("GET", "/token?code="+redisKey, nil)
	expect.Ok(t, err)

	response = testRouter(config, request)
	expecthttp.Status(t, http.StatusBadRequest, response)
	expecthttp.StringBody(t, "bad user id\n", response)

	// without code parameter
	request, err = http.NewRequest("GET", "/token", nil)
	expect.Ok(t, err)

	response = testRouter(config, request)
	expecthttp.Status(t, http.StatusBadRequest, response)
	expecthttp.StringBody(t, "code not set\n", response)
}

func TestTestMiddleware(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	recorder := httptest.NewRecorder()
	claims := &auth.Claims{UserID: "user-id"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		claim, ok := auth.UserFromContext(r.Context())
		expect.True(t, ok)
		expect.Equal(t, claims, claim)
	}

	router := mux.NewRouter()
	router.Use(auth.TestMiddleware(claims))
	router.Handle("/", http.HandlerFunc(handler)).Methods("GET")

	router.ServeHTTP(recorder, request)

	expecthttp.Ok(t, recorder)
}

func TestBuildHandler(t *testing.T) {
	t.Parallel()

	expectedClaims := &auth.Claims{UserID: uuid.New().String()}
	handler := auth.BuildHandler(func(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
		expect.Equal(t, expectedClaims, claims)

		return nil
	})

	router := mux.NewRouter()
	router.Use(auth.TestMiddleware(expectedClaims))
	router.Handle("/", handler).Methods("GET")
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	router.ServeHTTP(recorder, request)
	expecthttp.Ok(t, recorder)

	// Without Claim
	router = mux.NewRouter()
	router.Handle("/", handler).Methods("GET")
	recorder = httptest.NewRecorder()
	request, err = http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	router.ServeHTTP(recorder, request)
	expecthttp.Status(t, http.StatusUnauthorized, recorder)
}
