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

func authenticate(c *Config, r *http.Request) (*User, error) {
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

	// TODO: do we need to get the user?
	user, err := c.userStore.Get(claims.UserId)
	if err != nil {
		return nil, fmt.Errorf("error fetching user (%v) %v", claims.UserId, err)
	}

	return user, nil
}

func (c *Config) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := authenticate(c, r)
		if err != nil {
			code := http.StatusUnauthorized

			w.Header().Set("WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`)
			http.Error(w, http.StatusText(code), code)

			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		beeline.AddField(ctx, "user_id", user.Id)

		next.ServeHTTP(w, r.WithContext(ctx))

		if token, err := c.SignedToken(user); err == nil {
			w.Header().Set("Bissy-Token", token)
		}
	})
}
