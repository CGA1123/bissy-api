package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

func testConfig(t *testing.T, now time.Time, userID, redisID string, client utils.HTTPClient) (*auth.Config, *auth.SQLUserStore, *auth.RedisStateStore, func()) {
	db, dbTeardown := utils.TestDB(t)
	redisClient, redisTeardown := utils.TestRedis(t)
	redis := &auth.RedisStateStore{Client: redisClient, IDGenerator: &utils.TestIDGenerator{ID: redisID}}
	githubApp := auth.NewGithubApp("client-id", "client-secret", client)

	store := testSQLUserStore(now.Truncate(time.Millisecond), userID, hnysqlx.WrapDB(db))
	signingKey := []byte("test-key")
	clock := &utils.TestClock{Time: now}
	config := auth.NewConfig(
		signingKey,
		store,
		clock,
		redis,
		githubApp,
	)

	return config, store, redis, func() {
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

	c.Middleware(router).ServeHTTP(recorder, r)

	return recorder
}

func TestAuthHandler(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Millisecond)
	userID := uuid.New().String()
	config, store, _, teardown := testConfig(
		t,
		now,
		userID,
		uuid.New().String(),
		utils.NewTestHTTPClient(),
	)
	defer teardown()

	user, err := store.Create(&auth.CreateUser{GithubID: "github-id", Name: "Chritian"})
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	r := testHandler(config, request, testingHandler(t, user))
	expecthttp.Status(t, http.StatusUnauthorized, r)
	expecthttp.Header(t, "WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`, r.Header())

	// with bad auth header
	request, err = http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)
	request.Header.Add("Authorization", "Bearer 123453")

	r = testHandler(config, request, testingHandler(t, user))
	expecthttp.Status(t, http.StatusUnauthorized, r)
	expecthttp.Header(t, "WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`, r.Header())

	// with correct auth header
	token, err := config.SignedToken(user)
	expect.Ok(t, err)

	request, err = http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)
	request.Header.Add("Authorization", "Bearer "+token)
	r = testHandler(config, request, testingHandler(t, user))
	expecthttp.Ok(t, r)
}

func TestTokenHandler(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	userID := uuid.New().String()
	redisID := uuid.New().String()
	httpClient := utils.NewTestHTTPClient()
	config, store, redis, teardown := testConfig(t, now, userID, redisID, httpClient)
	defer teardown()

	user, err := store.Create(&auth.CreateUser{GithubID: "github-id", Name: "Chritian"})
	expect.Ok(t, err)

	_, err = redis.Set(user.ID, time.Minute*5)
	expect.Ok(t, err)

	token, err := config.SignedToken(user)
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/token?code="+redisID, nil)
	expect.Ok(t, err)

	response := testRouter(config, request)
	expecthttp.Ok(t, response)
	expecthttp.JSONBody(t, map[string]interface{}{"token": token}, response.Body)
	expecthttp.ContentType(t, handlerutils.ContentTypeJSON, response)
}
