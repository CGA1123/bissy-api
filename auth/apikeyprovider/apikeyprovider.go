package apikeyprovider

import (
	"net/http"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/auth/apikey"
)

// HeaderKey is the HTTP Header name expected to be populated by this provider
const HeaderKey = "x-bissy-apikey"

// Config holds the configuration for providing api key based authentication
// It implements the auth.Provider interface
type Config struct {
	store apikey.Store
}

// New configures a apikeyprovider backed with the given apikey.Store
func New(store apikey.Store) *Config {
	return &Config{store: store}
}

// Valid checks whether a given request is attempting apikey authentication
func (c *Config) Valid(r *http.Request) bool {
	return r.Header.Get(HeaderKey) != ""
}

// Authenticate attempts to authenticate a request
func (c *Config) Authenticate(r *http.Request) (*auth.Claims, bool) {
	key, err := c.store.GetByKey(r.Header.Get(HeaderKey))
	if err != nil {
		return nil, false
	}

	return &auth.Claims{UserID: key.UserID}, true
}
