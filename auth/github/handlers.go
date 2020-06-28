package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/auth/jwtprovider"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

// Config contains all the values required to support the auth package
type Config struct {
	jwt       *jwtprovider.Config
	userStore auth.UserStore
	clock     utils.Clock
	redis     StateStore
	githubApp *App
}

// TestConfig builds a config used for testing
func TestConfig(jwtConfig *jwtprovider.Config, store auth.UserStore, stateStore StateStore, githubApp *App, now time.Time) *Config {
	return &Config{
		jwt:       jwtConfig,
		userStore: store,
		clock:     &utils.TestClock{Time: now},
		redis:     stateStore,
		githubApp: githubApp,
	}
}

// New build a new Config struct
func New(jwtConfig *jwtprovider.Config, db *hnysqlx.DB, client *redis.Client, githubApp *App) *Config {
	return &Config{
		jwt:       jwtConfig,
		userStore: auth.NewSQLUserStore(db),
		clock:     &utils.RealClock{},
		redis:     &RedisStateStore{Client: client, IDGenerator: &utils.UUIDGenerator{}},
		githubApp: githubApp,
	}
}

// SetupHandlers mounts the auth HTTP handlers on the given router
func (c *Config) SetupHandlers(router *mux.Router) {
	router.
		Handle("/signin", &handlerutils.Handler{H: c.signin}).
		Methods("GET")

	router.
		Handle("/callback", &handlerutils.Handler{H: c.callback}).
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
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("bad user id"), Status: http.StatusBadRequest}
	}

	token, err := c.jwt.SignedToken(user)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("error signing token"), Status: http.StatusInternalServerError}
	}

	handlerutils.ContentType(w, handlerutils.ContentTypeJSON)
	return json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{Token: token})
}

func (c *Config) signin(w http.ResponseWriter, r *http.Request) error {
	redirectURL, ok := handlerutils.Params(r).Get("redirect_uri")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("redirect_uri not set"), Status: http.StatusBadRequest}
	}

	reader, err := json.Marshal(&ClientState{Redirect: redirectURL})
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("error marshalling state"), Status: http.StatusInternalServerError}
	}

	state, err := c.redis.Set(string(reader), 5*time.Minute)
	if err != nil {
		return err
	}

	githubURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%v&state=%v&scope=user",
		c.githubApp.clientID,
		state)

	http.Redirect(w, r, githubURL, http.StatusTemporaryRedirect)

	return nil
}

func (c *Config) callback(w http.ResponseWriter, r *http.Request) error {
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

	clientState, err := c.fetchState(state)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("error fetching state set"), Status: http.StatusBadRequest}
	}

	oauth, err := c.githubApp.OAuthClient(code, state)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("could not fetch token"), Status: http.StatusBadRequest}
	}

	createUser, err := oauth.User()
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("could not fetch github user"), Status: http.StatusBadRequest}
	}

	user, err := getOrCreateUser(c.userStore, createUser)
	if err != nil {
		return err
	}

	code, err = c.redis.Set(user.ID, time.Minute*5)
	if err != nil {
		return fmt.Errorf("error setting code: %v", err)
	}

	redirectURL := fmt.Sprintf("%v?code=%v", clientState.Redirect, code)

	http.Redirect(w, r, redirectURL, http.StatusFound)

	return nil
}
