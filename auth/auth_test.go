package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/auth/jwtprovider"
	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/expect"
	"github.com/cga1123/bissy-api/utils/expecthttp"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func testingHandler(t *testing.T, expected *auth.User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claim, ok := auth.UserFromContext(r.Context())

		expect.True(t, ok)
		expect.Equal(t, expected.ID, claim.UserID)
		expect.Equal(t, expected.Name, claim.Name)
	})
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

func testConfig(t *testing.T, now time.Time, provider auth.Provider, userID string) (*auth.Auth, *auth.SQLUserStore, func()) {
	db, dbTeardown := utils.TestDB(t)

	store := auth.TestSQLUserStore(now.Truncate(time.Millisecond), userID, db)
	config := &auth.Auth{Providers: []auth.Provider{provider}}

	return config, store, func() {
		expect.Ok(t, dbTeardown())
	}
}

// TODO: move JWT specific tests to jwtprovider & make tests more generic using
// Provider interface
func TestAuthHandler(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Millisecond)
	signingKey := []byte("test-key")
	jwtProvider := jwtprovider.TestConfig(signingKey, now)

	userID := uuid.New().String()
	authConfig, store, teardown := testConfig(
		t,
		now,
		jwtProvider,
		userID,
	)
	defer teardown()

	user, err := store.Create(&auth.CreateUser{GithubID: "github-id", Name: "Chritian"})
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)

	r := testHandler(authConfig, request, testingHandler(t, user))
	expecthttp.Status(t, http.StatusUnauthorized, r)
	expecthttp.Header(t, "WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`, r.Header())

	// with bad auth header
	request, err = http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)
	request.Header.Add("Authorization", "Bearer 123453")

	r = testHandler(authConfig, request, testingHandler(t, user))
	expecthttp.Status(t, http.StatusUnauthorized, r)
	expecthttp.Header(t, "WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`, r.Header())

	// with correct auth header
	token, err := jwtProvider.SignedToken(user)
	expect.Ok(t, err)

	request, err = http.NewRequest("GET", "/", nil)
	expect.Ok(t, err)
	request.Header.Add("Authorization", "Bearer "+token)
	r = testHandler(authConfig, request, testingHandler(t, user))
	expecthttp.Ok(t, r)
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
