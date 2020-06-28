package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cga1123/bissy-api/utils/handlerutils"
	"github.com/honeycombio/beeline-go"
)

// The Provider interface describes and authentication provider, it contains two
// methods:
// - Valid which checks whether the given request is attempting to authenticate with a give provider
// - Authenticate which attempts to authenticate the request
type Provider interface {
	Valid(*http.Request) bool
	Authenticate(*http.Request) (*Claims, bool)
}

type contextKey int

const (
	userContextKey contextKey = iota
)

// Claims represents the custom JWT Claims struct
type Claims struct {
	UserID string `json:"user_id"`
	Name   string
}

// Auth contains the signing key for generation signed tokens
type Auth struct {
	Providers []Provider
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
func (c *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claim, ok := c.Provider(r)
		if !ok {
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

// Provider returns a matching provider for a request, if any
func (c *Auth) Provider(r *http.Request) (*Claims, bool) {
	for _, provider := range c.Providers {
		if provider.Valid(r) {
			return provider.Authenticate(r)
		}
	}

	return nil, false
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

// BuildHandler builds a handlerutils.Handler and passes in the *Claims if set,
// returns http.StatusUnauthorized if not.
func BuildHandler(next func(*Claims, http.ResponseWriter, *http.Request) error) http.Handler {
	return &handlerutils.Handler{H: func(w http.ResponseWriter, r *http.Request) error {
		claim, ok := UserFromContext(r.Context())
		if !ok {
			return &handlerutils.HandlerError{Status: http.StatusUnauthorized, Err: fmt.Errorf("no claim present")}
		}

		return next(claim, w, r)
	}}
}
