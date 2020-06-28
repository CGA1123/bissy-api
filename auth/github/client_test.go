package github_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/auth/github"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/expecthttp"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
)

func mockGithubUserFetch(t *testing.T, id, name string, client *utils.TestHTTPClient) {
	expectedBody := map[string]interface{}{"query": "query { viewer { id, name } }"}

	client.Mock("POST", "https://api.github.com/graphql", func(r *http.Request) (*http.Response, error) {
		var body interface{}
		if r.Body == nil {
			t.Fatalf("request body is nil")
		}

		err := utils.ParseJSONBody(r.Body, &body)
		expect.Ok(t, err)

		expecthttp.Header(t, "Authorization", "token my-access-token", r.Header)
		expect.Equal(t, expectedBody, body)

		jsonString := fmt.Sprintf(`
			{
				"data": {
					"viewer": {
						"id": "%v",
						"name": "%v"
					}
				}
			}
		`, id, name)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(jsonString)),
		}, nil
	})
}

func mockGithubTokenExchange(t *testing.T, client *utils.TestHTTPClient, code, state string) {
	expectedURL := "https://api.github.com/login/oauth/access_token"
	expectedBody := map[string]interface{}{
		"client_id":     "client-id",
		"client_secret": "client-secret",
		"code":          code,
		"state":         state,
	}

	client.Mock("POST", expectedURL, func(r *http.Request) (*http.Response, error) {
		var body interface{}
		err := utils.ParseJSONBody(r.Body, &body)
		expect.Ok(t, err)

		expecthttp.Header(t, "Accept", handlerutils.ContentTypeJSON, r.Header)
		expect.Equal(t, expectedBody, body)
		jsonString := `{
			"access_token":"my-access-token",
			"scope":"user",
			"token_type":"bearer"}`

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(jsonString)),
		}, nil
	})
}

func TestOAuthClient(t *testing.T) {
	t.Parallel()

	client := utils.NewTestHTTPClient()
	app := github.NewApp("client-id", "client-secret", client)

	mockGithubTokenExchange(t, client, "my-code", "my-state")

	oauth, err := app.OAuthClient("my-code", "my-state")
	expect.Ok(t, err)
	expect.Equal(t, "my-access-token", oauth.Token)
	expect.Equal(t, "https://api.github.com", oauth.Base)
}

func TestOAuthDo(t *testing.T) {
	t.Parallel()

	client := utils.NewTestHTTPClient()
	oauth := &github.OAuthClient{
		Base: "https://example.com", Token: "my-access-token", HTTPClient: client}

	request, err := http.NewRequest("PURGE", "https://example.com/hello", nil)
	expect.Ok(t, err)

	mockedHandler := func(r *http.Request) (*http.Response, error) {
		expecthttp.Header(t, "Authorization", "token my-access-token", r.Header)

		return &http.Response{StatusCode: http.StatusOK}, nil
	}

	client.Mock("PURGE", "https://example.com/hello", mockedHandler)
	_, err = oauth.Do(request)
	expect.Ok(t, err)
}

func TestOAuthUser(t *testing.T) {
	t.Parallel()
	client := utils.NewTestHTTPClient()
	oauth := &github.OAuthClient{
		Base: "https://api.github.com", Token: "my-access-token", HTTPClient: client}
	expectedUser := &auth.CreateUser{GithubID: "github-user-id", Name: "Bissy"}

	mockGithubUserFetch(t, "github-user-id", "Bissy", client)

	user, err := oauth.User()
	expect.Ok(t, err)
	expect.Equal(t, expectedUser, user)

	// error url parsing
	oauth = &github.OAuthClient{
		Base: "\000api.com", Token: "my-access-token", HTTPClient: client}

	mockGithubUserFetch(t, "github-user-id", "Bissy", client)

	_, err = oauth.User()
	expect.Error(t, err)
	expect.Equal(t, "error building request: parse \"\\x00api.com/graphql\": net/url: invalid control character in URL", err.Error())

	// error on request
	oauth = &github.OAuthClient{
		Base: "https://api.github.com", Token: "my-access-token", HTTPClient: client}
	client.Mock("POST", "https://api.github.com/graphql", func(r *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("errored!")
	})
	_, err = oauth.User()
	expect.Error(t, err)
	expect.Equal(t, "error doing request: errored!", err.Error())

	// error parsing response
	client.Mock("POST", "https://api.github.com/graphql", func(r *http.Request) (*http.Response, error) {
		response := httptest.NewRecorder()
		return response.Result(), nil
	})
	_, err = oauth.User()
	expect.Error(t, err)
	expect.Equal(t, "error parsing response: EOF", err.Error())
}

func TestGithubSignIn(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	userID := uuid.New().String()
	redisID := uuid.New().String()
	config, _, _, redis, teardown := testConfig(
		t, now, userID, redisID, utils.NewTestHTTPClient())
	defer teardown()

	request, err := http.NewRequest("GET",
		"/signin?redirect_uri=https://app.bissy.io", nil)
	expect.Ok(t, err)

	response := testRouter(config, request)

	githubRedirectUrl := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%v&state=%v&scope=user",
		"client-id", redisID)
	expecthttp.Status(t, http.StatusTemporaryRedirect, response)
	expecthttp.Header(t, "Location", githubRedirectUrl, response.Header())

	exists, err := redis.Exists(redisID)
	expect.Ok(t, err)
	expect.True(t, exists)

	clientState, err := redis.Get(redisID)
	expect.Ok(t, err)
	expecthttp.JSONBody(
		t,
		&github.ClientState{Redirect: "https://app.bissy.io"},
		bytes.NewBuffer([]byte(clientState)))
}

func TestGithubCallback(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	userID := uuid.New().String()
	redisID := uuid.New().String()
	httpClient := utils.NewTestHTTPClient()
	config, _, store, redis, teardown := testConfig(t, now, userID, redisID, httpClient)
	defer teardown()

	clientState := `{"redirect": "https://app.bissy.io"}`
	_, err := redis.Set(clientState, time.Hour)
	expect.Ok(t, err)

	expectedUser := &auth.User{ID: userID, GithubID: "github-user-id", Name: "Bissy", CreatedAt: now}
	expect.Ok(t, err)

	request, err := http.NewRequest("GET", "/callback?code=my-code&state="+redisID, nil)
	expect.Ok(t, err)

	mockGithubTokenExchange(t, httpClient, "my-code", redisID)
	mockGithubUserFetch(t, "github-user-id", "Bissy", httpClient)

	response := testRouter(config, request)

	expecthttp.Status(t, http.StatusFound, response)

	// creates user
	user, err := store.Get(userID)
	expect.Ok(t, err)

	// response
	expectedRedirect := fmt.Sprintf("https://app.bissy.io?code=%v", redisID)
	expecthttp.Status(t, http.StatusFound, response)
	expect.Equal(t, expectedUser, user)
	expecthttp.Header(t, "Location", expectedRedirect, response.Header())
	userID, err = redis.Get(redisID)
	expect.Ok(t, err)
	expect.Equal(t, user.ID, userID)

	// when user already exists
	_, err = redis.Set(clientState, time.Hour)
	expect.Ok(t, err)

	response = testRouter(config, request)
	expecthttp.Status(t, http.StatusFound, response)
	expecthttp.Header(t, "Location", expectedRedirect, response.Header())
	userID, err = redis.Get(redisID)
	expect.Ok(t, err)
	expect.Equal(t, user.ID, userID)

	// without code
	request, err = http.NewRequest("GET", "/callback?state="+redisID, nil)
	expect.Ok(t, err)
	response = testRouter(config, request)
	expecthttp.Status(t, http.StatusBadRequest, response)

	// without state
	request, err = http.NewRequest("GET", "/callback?code=my-code", nil)
	expect.Ok(t, err)
	response = testRouter(config, request)
	expecthttp.Status(t, http.StatusBadRequest, response)

}
