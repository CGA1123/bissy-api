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
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type BissyClaims struct {
	UserId string `json:"user_id"`
	jwt.StandardClaims
}

type CacheStore interface {
	Exists(string) (bool, error)
	Set(string, time.Duration) error
	Del(string) (bool, error)
}

type RedisStore struct {
	Client *redis.Client
}

func (r *RedisStore) Exists(key string) (bool, error) {
	value, err := r.Client.Exists(context.TODO(), "auth:"+key).Result()
	if err != nil {
		return false, err
	}

	return value == 1, nil
}

func (r *RedisStore) Set(key string, exp time.Duration) error {
	_, err := r.Client.Set(
		context.TODO(),
		"auth:"+key,
		key,
		exp,
	).Result()

	return err
}

func (r *RedisStore) Del(key string) (bool, error) {
	count, err := r.Client.Del(context.TODO(), key).Result()
	if err != nil {
		return false, err
	}

	return count == 1, nil
}

type Config struct {
	signingKey     []byte
	expiryTime     time.Duration
	userStore      UserStore
	clock          utils.Clock
	redis          CacheStore
	idGenerator    utils.IdGenerator
	githubClientId string
}

func NewConfig(key []byte, duration time.Duration, store UserStore, clock utils.Clock, redis CacheStore, gen utils.IdGenerator, githubClientId string) *Config {
	return &Config{
		signingKey:     key,
		expiryTime:     duration,
		userStore:      store,
		clock:          clock,
		redis:          redis,
		idGenerator:    gen,
		githubClientId: githubClientId,
	}
}

func (c *Config) SignedToken(u *User) (string, error) {
	token := u.NewToken(c.clock.Now().Add(c.expiryTime))

	return token.SignedString(c.signingKey)
}

func (c *Config) githubSignup(code string) (*User, error) {
	return nil, fmt.Errorf("nyi")
}

func authenticate(c *Config, r *http.Request) (*User, error) {
	header := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(header) != 2 {
		return nil, fmt.Errorf("bad Authorization header (%v)", r.Header.Get("Authorization"))
	}

	token, err := jwt.ParseWithClaims(header[1], &BissyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return c.signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*BissyClaims)
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

		next.ServeHTTP(w, r.WithContext(ctx))

		if token, err := c.SignedToken(user); err == nil {
			w.Header().Set("Bissy-Token", token)
		}
	})
}

func (c *Config) SetupHandlers(router *mux.Router) {
	router.
		Handle("/github/signin", &handlerutils.Handler{H: c.githubSignin}).
		Methods("GET")

	router.
		Handle("/github/callback", &handlerutils.Handler{H: c.githubCallback}).
		Methods("GET")
}
