package github

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/handlerutils"
)

const baseURL = "https://api.github.com"

// ClientState represents the state passed in by a client auth request,
// including a random State string, and the Redirect URI callback
type ClientState struct {
	Redirect string
}

// App represents the configuration of the Github OAuth application used
// to authenticate users.
type App struct {
	clientID     string
	clientSecret string
	httpClient   utils.HTTPClient
}

// OAuthClient is a http client that automatically adds the required
// Authorization headers when called
type OAuthClient struct {
	Token      string
	HTTPClient utils.HTTPClient
	Base       string
}

// Do executes a request using OAuthClient
func (c *OAuthClient) Do(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", "token "+c.Token)

	return c.HTTPClient.Do(r)
}

// User fetches the currently authenticated user from  and returns the
// struct to be passed to create the user.
func (c *OAuthClient) User() (*auth.CreateUser, error) {
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

	return &auth.CreateUser{
		GithubID: body.Data.Viewer.ID,
		Name:     body.Data.Viewer.Name,
	}, nil
}

// NewApp configures a new GithubApp struct
func NewApp(id, secret string, client utils.HTTPClient) *App {
	return &App{clientID: id, clientSecret: secret, httpClient: client}
}

// OAuthClient swaps a code token for a OAuthClient to make subsequent
// authenticated requests to the  API
func (ga *App) OAuthClient(code, state string) (*OAuthClient, error) {
	request, err := ga.buildOauthRequest(code, state)
	if err != nil {
		return nil, err
	}

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

	return &OAuthClient{
		Token:      tokenStruct.Token,
		Base:       baseURL,
		HTTPClient: ga.httpClient,
	}, nil
}

func (ga *App) buildOauthRequest(code, state string) (*http.Request, error) {
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

	return request, nil
}

func getOrCreateUser(store auth.UserStore, cu *auth.CreateUser) (*auth.User, error) {
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
