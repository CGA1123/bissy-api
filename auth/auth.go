package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BissyClaims struct {
	UserId string `json:"user_id"`
	jwt.StandardClaims
}

type User struct {
	Id        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type CreateUser struct {
	Email string
}

type UserStore interface {
	Get(string) (*User, error)
	Create(*CreateUser) (*User, error)
}

func (u *User) NewToken(exp time.Time) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, BissyClaims{
		u.Id,
		jwt.StandardClaims{
			ExpiresAt: exp.Unix(),
			Issuer:    "bissy-api",
		},
	})
}

type IdGenerator interface {
	Generate() string
}

type Clock interface {
	Now() time.Time
}

type UUIDGenerator struct{}

func (generator *UUIDGenerator) Generate() string {
	return uuid.New().String()
}

type SQLUserStore struct {
	db          *sqlx.DB
	idGenerator IdGenerator
	clock       Clock
}

func NewSQLUserStore(db *sqlx.DB, clock Clock, gen IdGenerator) *SQLUserStore {
	return &SQLUserStore{db: db, idGenerator: gen, clock: clock}
}

func (s *SQLUserStore) Get(id string) (*User, error) {
	var user User

	query := "SELECT * FROM auth_users WHERE id = $1"

	if err := s.db.Get(&user, query, id); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SQLUserStore) Create(cu *CreateUser) (*User, error) {
	now := s.clock.Now()
	id := s.idGenerator.Generate()

	query := `
		INSERT INTO auth_users (id, email, created_at)
		VALUES ($1, $2, $3)
		RETURNING *`

	var user User
	if err := s.db.Get(&user, query, id, cu.Email, now); err != nil {
		return nil, err
	}

	return &user, nil
}

type Config struct {
	signingKey []byte
	expiryTime time.Duration
	userStore  UserStore
	clock      Clock
}

func NewConfig(key []byte, duration time.Duration, store UserStore, clock Clock) *Config {
	return &Config{signingKey: key, expiryTime: duration, userStore: store, clock: clock}
}

func (c *Config) SignedToken(u *User) (string, error) {
	token := u.NewToken(c.clock.Now().Add(c.expiryTime))

	return token.SignedString(c.signingKey)
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

			http.Error(w, http.StatusText(code), code)
			w.Header().Set("WWW-Authenticate", `Bearer realm="bissy-api" charset="UTF-8"`)

			return
		}

		ctx := context.WithValue(r.Context(), "user", user)

		next.ServeHTTP(w, r.WithContext(ctx))

		if token, err := c.SignedToken(user); err == nil {
			w.Header().Set("Bissy-Token", token)
		}
	})
}
