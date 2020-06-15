package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/honeycombio/beeline-go"
)

type contextKey int

const (
	userContextKey contextKey = iota
)

// Claims represents the custom JWT Claims struct
type Claims struct {
	UserID string `json:"user_id"`
	Name   string
	jwt.StandardClaims
}

// Config contains all the values required to support the auth package
type Config struct {
	signingKey []byte
	expiryTime time.Duration
	userStore  UserStore
	clock      utils.Clock
	redis      StateStore
	githubApp  *GithubApp
}

// NewConfig build a new Config struct
func NewConfig(key []byte, store UserStore, clock utils.Clock, redis StateStore, githubApp *GithubApp) *Config {
	return &Config{
		signingKey: key,
		userStore:  store,
		clock:      clock,
		redis:      redis,
		githubApp:  githubApp,
	}
}

// SignedToken returns a new signed JWT token string for the given User
func (c *Config) SignedToken(u *User) (string, error) {
	token := u.NewToken(c.clock.Now().Add(12 * time.Hour))

	return token.SignedString(c.signingKey)
}

// UserFromContext fetches the Claim from the current context
func UserFromContext(ctx context.Context) (*Claims, bool) {
	claim, ok := ctx.Value(userContextKey).(*Claims)

	return claim, ok
}

// Middleware ensures that a request is authentic before passing the request on
// to the next middleware in the stack.
// It will inject a Claim struct into the request context on successful authentication
// This Claim can be retrieved in downsteam handlers via UserFromContext.
func (c *Config) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claim, err := authenticate(c, r)
		if err != nil {
			code := http.StatusUnauthorized

			w.Header().Set("WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`)
			http.Error(w, http.StatusText(code), code)

			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claim)
		beeline.AddField(ctx, "user_id", claim.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TestMiddleware will return a middleware which injects the given claim into
// requests.
func TestMiddleware(claim *Claims) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), userContextKey, claim)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// BuildHandler builds a ErrorHandler and passes in the *Claims if set,
// returns http.StatusUnauthorized if not.
func BuildHandler(next func(*Claims, http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		claim, ok := UserFromContext(r.Context())
		if !ok {
			return &handlerutils.HandlerError{Status: http.StatusUnauthorized, Err: fmt.Errorf("no claim present")}
		}

		return next(claim, w, r)
	}
}

func authenticate(c *Config, r *http.Request) (*Claims, error) {
	header := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(header) != 2 {
		return nil, fmt.Errorf("bad Authorization header (%v)", r.Header.Get("Authorization"))
	}

	token, err := jwt.ParseWithClaims(header[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return c.signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("could not verify token")
	}

	return claims, nil
}
