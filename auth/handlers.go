package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/gorilla/mux"
)

// SetupHandlers mounts the auth HTTP handlers on the given router
func (c *Config) SetupHandlers(router *mux.Router) {
	router.
		Handle("/github/signin", &handlerutils.Handler{H: c.githubSignin}).
		Methods("GET")

	router.
		Handle("/github/callback", &handlerutils.Handler{H: c.githubCallback}).
		Methods("GET")

	router.
		Handle("/token", &handlerutils.Handler{H: c.token}).
		Methods("GET")
}

func (c *Config) token(w http.ResponseWriter, r *http.Request) error {
	code, ok := handlerutils.Params(r).Get("code")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("code not set"), Status: http.StatusBadRequest}
	}

	userID, err := c.redis.Get(code)
	if err != nil || userID == "" {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("bad token"), Status: http.StatusBadRequest}
	}

	user, err := c.userStore.Get(userID)
	if err != nil || userID == "" {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("bad user id"), Status: http.StatusBadRequest}
	}

	token, err := c.SignedToken(user)
	if err != nil || userID == "" {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("error signing token"), Status: http.StatusInternalServerError}
	}

	handlerutils.ContentType(w, handlerutils.ContentTypeJSON)
	return json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{Token: token})
}
