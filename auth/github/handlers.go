package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/auth/jwtprovider"
	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/cache"
	"github.com/cga1123/bissy-api/utils/handlerutils"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

// Config contains all the values required to support the auth package
type Config struct {
	jwt       *jwtprovider.Config
	userStore auth.UserStore
	clock     utils.Clock
	redis     cache.StateStore
	githubApp *App
}

// TestConfig builds a config used for testing
func TestConfig(jwtConfig *jwtprovider.Config, store auth.UserStore, stateStore cache.StateStore, githubApp *App, now time.Time) *Config {
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
		redis:     &cache.RedisStateStore{Client: client, IDGenerator: &utils.UUIDGenerator{}, Prefix: "github"},
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

	user, err := c.getUser(code)
	if err != nil {
		return err
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

func (c *Config) getUser(code string) (*auth.User, error) {
	userID, err := c.redis.Get(code)
	if err != nil || userID == "" {
		return nil, &handlerutils.HandlerError{
			Err: fmt.Errorf("bad token"), Status: http.StatusBadRequest}
	}

	user, err := c.userStore.Get(userID)
	if err != nil {
		return nil, &handlerutils.HandlerError{
			Err: fmt.Errorf("bad user id"), Status: http.StatusBadRequest}
	}

	return user, nil
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
	requiredParams, err := requireParams(r, "code", "state")
	if err != nil {
		return err
	}

	code := requiredParams["code"]
	state := requiredParams["state"]

	clientState, err := c.fetchState(state)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("error fetching state set"), Status: http.StatusBadRequest}
	}

	user, err := createUser(c, code, state)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	code, err = c.redis.Set(user.ID, time.Minute*5)
	if err != nil {
		return fmt.Errorf("error setting code: %v", err)
	}

	redirectURL := fmt.Sprintf("%v?code=%v", clientState.Redirect, code)

	http.Redirect(w, r, redirectURL, http.StatusFound)

	return nil
}

func createUser(c *Config, code, state string) (*auth.User, error) {
	oauth, err := c.githubApp.OAuthClient(code, state)
	if err != nil {
		return nil, &handlerutils.HandlerError{
			Err: fmt.Errorf("could not fetch token"), Status: http.StatusBadRequest}
	}

	createUser, err := oauth.User()
	if err != nil {
		return nil, &handlerutils.HandlerError{
			Err: fmt.Errorf("could not fetch github user"), Status: http.StatusBadRequest}
	}

	return getOrCreateUser(c.userStore, createUser)
}

func requireParams(r *http.Request, required ...string) (map[string]string, error) {
	params := handlerutils.Params(r)
	requiredParams := map[string]string{}

	for _, param := range required {
		value, ok := params.Get(param)
		if !ok {
			return requiredParams, &handlerutils.HandlerError{
				Err: fmt.Errorf("%v not set", param), Status: http.StatusBadRequest}
		}

		requiredParams[param] = value
	}

	return requiredParams, nil
}
