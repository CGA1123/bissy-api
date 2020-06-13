package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
)

// ClientState represents the state passed in by a client auth request,
// including a random State string, and the Redirect URI callback
type ClientState struct {
	State    string
	Redirect string
}

// GithubApp represents the configuration of the Github OAuth application used
// to authenticate users.
type GithubApp struct {
	clientID     string
	clientSecret string
	httpClient   utils.HTTPClient
}

// GithubOAuthClient is a http client that automatically adds the required
// Authorization headers when called
type GithubOAuthClient struct {
	Token      string
	HTTPClient utils.HTTPClient
	Base       string
}

// Do executes a request using GithubOAuthClient
func (c *GithubOAuthClient) Do(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", "token "+c.Token)

	return c.HTTPClient.Do(r)
}

// User fetches the currently authenticated user from Github and returns the
// struct to be passed to create the user.
func (c *GithubOAuthClient) User() (*CreateUser, error) {
	json := strings.NewReader(`{ "query": "query { viewer { id, name } }" }`)
	request, err := http.NewRequest("POST", c.Base+"/graphql", json)
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}

	response, err := c.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %v", err)
	}

	var body struct {
		Data struct {
			Viewer struct {
				ID   string
				Name string
			}
		}
	}

	if err := utils.ParseJSONBody(response.Body, &body); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &CreateUser{
		GithubID: body.Data.Viewer.ID,
		Name:     body.Data.Viewer.Name,
	}, nil
}

// NewGithubApp configures a new GithubApp struct
func NewGithubApp(id, secret string, client utils.HTTPClient) *GithubApp {
	return &GithubApp{clientID: id, clientSecret: secret, httpClient: client}
}

// OAuthClient swaps a code token for a GithubOAuthClient to make subsequent
// authenticated requests to the Github API
func (ga *GithubApp) OAuthClient(code, state string) (*GithubOAuthClient, error) {
	baseURL := "https://api.github.com"
	body, err := utils.JSONBody(map[string]string{
		"client_id":     ga.clientID,
		"client_secret": ga.clientSecret,
		"code":          code,
		"state":         state,
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", baseURL+"/login/oauth/access_token", body)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Accept", handlerutils.ContentTypeJSON)

	response, err := ga.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	var tokenStruct struct {
		Token string `json:"access_token"`
	}

	if err := utils.ParseJSONBody(response.Body, &tokenStruct); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &GithubOAuthClient{
		Token:      tokenStruct.Token,
		Base:       baseURL,
		HTTPClient: ga.httpClient,
	}, nil
}

func (c *Config) githubSignin(w http.ResponseWriter, r *http.Request) error {
	redirectURL, ok := handlerutils.Params(r).Get("redirect_uri")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("redirect_uri not set"), Status: http.StatusBadRequest}
	}

	clientState, ok := handlerutils.Params(r).Get("state")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("state not set"), Status: http.StatusBadRequest}
	}

	reader, err := json.Marshal(&ClientState{State: clientState, Redirect: redirectURL})
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

func getOrCreateUser(store UserStore, cu *CreateUser) (*User, error) {
	user, err := store.GetByGithubID(cu.GithubID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == nil {
		return user, nil
	}

	return store.Create(cu)
}

func (c *Config) fetchState(key string) (*ClientState, error) {
	ok, err := c.redis.Exists(key)
	if !ok || err != nil {
		return nil, fmt.Errorf("failed to check key exists: %v", err)
	}

	clientStateJSON, err := c.redis.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %v", err)
	}

	var clientState ClientState
	if err := utils.ParseJSONBody(ioutil.NopCloser(strings.NewReader(clientStateJSON)), &clientState); err != nil {
		return nil, fmt.Errorf("error parsing client state: %v", err)
	}

	// TODO: Delete state key...

	return &clientState, nil
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

	redirectURL := fmt.Sprintf("%v?code=%v&state=%v", clientState.Redirect, code, clientState.State)

	http.Redirect(w, r, redirectURL, http.StatusFound)

	return nil
}
