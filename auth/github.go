package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cga1123/bissy-api/handlerutils"
)

// swap code for oauth token
func githubToken(c *Config, code string) (string, error) {
	return "", fmt.Errorf("nyi")
}

func githubCreateUser(oauth string) (*CreateUser, error) {
	return nil, fmt.Errorf("nyi")
}

func (c *Config) githubSignin(w http.ResponseWriter, r *http.Request) error {
	state := c.idGenerator.Generate()
	if err := c.redis.Set(state, 5*time.Minute); err != nil {
		return err
	}

	githubUrl := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%v&state=%v&scope=user",
		c.githubClientId,
		state)

	http.Redirect(w, r, githubUrl, http.StatusTemporaryRedirect)

	return nil
}

func (c *Config) githubCallback(w http.ResponseWriter, r *http.Request) error {
	params := handlerutils.Params(r)
	code, ok := params.Get("code")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("code not set"), Status: http.StatusBadRequest}
	}

	state, ok := params.Get("state")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("state not set"), Status: http.StatusBadRequest}
	}

	ok, err := c.redis.Exists(state)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("code not set"), Status: http.StatusInternalServerError}
	}

	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("bad state parameter"), Status: http.StatusBadRequest}
	}

	// TODO: Swap for OAUTH token, get user, create user, redirect to app
	token, err := githubToken(c, code)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("could not fetch token"), Status: http.StatusBadRequest}
	}

	_, err = githubCreateUser(token)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("could not fetch github user"), Status: http.StatusBadRequest}
	}

	// check if user already exists?
	// redirect to app with jwt
	//user, err := c.userStore.Create(createUser)
	//if err != nil {
	//}

	return nil
}
