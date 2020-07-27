package jwtprovider

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/utils"
	"github.com/dgrijalva/jwt-go"
)

type jwtClaims struct {
	auth.Claims
	jwt.StandardClaims
}

// Config contains the internal configuration for the jwtprovider
type Config struct {
	signingKey []byte
	clock      utils.Clock
}

// TestConfig builds a new test Config
func TestConfig(key []byte, now time.Time) *Config {
	return &Config{
		signingKey: key,
		clock:      &utils.TestClock{Time: now},
	}
}

// New builds a new Config struct
func New(key []byte) *Config {
	return &Config{
		signingKey: key,
		clock:      &utils.RealClock{},
	}
}

// SignedToken returns a new signed JWT token string for the given User
func (c *Config) SignedToken(u *auth.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwtClaims{
		auth.Claims{u.ID, u.Name},
		jwt.StandardClaims{
			ExpiresAt: c.clock.Now().Add(12 * time.Hour).Unix(),
			Issuer:    "bissy-api",
		},
	})

	return token.SignedString(c.signingKey)
}

// Valid checks whether the request is a valid attemp at JWT auth
func (c *Config) Valid(r *http.Request) bool {
	header := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	return len(header) == 2
}

// Authenticate authenticates a requests via a JWT Bearer token, returning the
// associated Claims if if authentication succeed
func (c *Config) Authenticate(r *http.Request) (*auth.Claims, bool) {
	header := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(header) != 2 {
		return nil, false
	}

	token, err := c.parseToken(header[1])
	if err != nil {
		return nil, false
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, false
	}

	return toClaims(claims), true
}

func (c *Config) parseToken(header string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(header, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return c.signingKey, nil
	})
}

func toClaims(jwtClaim *jwtClaims) *auth.Claims {
	return &auth.Claims{UserID: jwtClaim.UserID, Name: jwtClaim.Name}
}
