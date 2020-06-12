package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cga1123/bissy-api/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/honeycombio/beeline-go"
)

const (
	contextKey = "auth_user_claim"
)

type Claims struct {
	UserId string `json:"user_id"`
	Name   string
	jwt.StandardClaims
}

type Config struct {
	signingKey []byte
	expiryTime time.Duration
	userStore  UserStore
	clock      utils.Clock
	redis      CacheStore
	githubApp  *GithubApp
}

func NewConfig(key []byte, store UserStore, clock utils.Clock, redis CacheStore, githubApp *GithubApp) *Config {
	return &Config{
		signingKey: key,
		userStore:  store,
		clock:      clock,
		redis:      redis,
		githubApp:  githubApp,
	}
}

func (c *Config) SignedToken(u *User) (string, error) {
	token := u.NewToken(c.clock.Now().Add(12 * time.Hour))

	return token.SignedString(c.signingKey)
}

func UserFromContext(ctx context.Context) (*Claims, bool) {
	claim, ok := ctx.Value(contextKey).(*Claims)

	return claim, ok
}

func (c *Config) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claim, err := authenticate(c, r)
		if err != nil {
			code := http.StatusUnauthorized

			w.Header().Set("WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`)
			http.Error(w, http.StatusText(code), code)

			return
		}

		ctx := context.WithValue(r.Context(), contextKey, claim)
		beeline.AddField(ctx, "user_id", claim.UserId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
