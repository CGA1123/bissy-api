package apikey_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/auth/apikey"
	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/expect"
	"github.com/cga1123/bissy-api/utils/expecthttp"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func testConfig(t *testing.T) (*apikey.Config, *auth.User, func()) {
	t.Helper()

	testDb, teardownDB := utils.TestDB(t)

	config := &apikey.Config{Store: apikey.NewSQLStore(testDb)}
	user, err := auth.NewSQLUserStore(testDb).Create(&auth.CreateUser{GithubID: "CGA1123", Name: "Christian"})
	expect.Ok(t, err)

	teardown := func() {
		expect.Ok(t, teardownDB())
	}

	return config, user, teardown
}

func testHandler(claims *auth.Claims, config *apikey.Config, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.Use(auth.TestMiddleware(claims))
	config.SetupHandlers(router)

	router.ServeHTTP(recorder, r)

	return recorder
}

func TestList(t *testing.T) {
	t.Parallel()

	config, user, teardown := testConfig(t)
	defer teardown()

	newAPIKey, err := config.Store.Create(user.ID, &apikey.Create{Name: "test key"})
	expect.Ok(t, err)

	claims := &auth.Claims{UserID: user.ID}
	request, err := http.NewRequest("GET", "/apikeys", nil)
	expect.Ok(t, err)

	response := testHandler(claims, config, request)
	expecthttp.Ok(t, response)
	expectedKey := &apikey.Struct{
		ID:        newAPIKey.ID,
		UserID:    user.ID,
		Name:      "test key",
		LastUsed:  newAPIKey.LastUsed,
		CreatedAt: newAPIKey.CreatedAt}

	expecthttp.JSONBody(t, []*apikey.Struct{expectedKey}, response.Body)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	config, user, teardown := testConfig(t)
	defer teardown()

	newAPIKey, err := config.Store.Create(user.ID, &apikey.Create{Name: "test key"})
	expect.Ok(t, err)

	claims := &auth.Claims{UserID: user.ID}
	request, err := http.NewRequest("DELETE", "/apikeys/"+newAPIKey.ID, nil)
	expect.Ok(t, err)

	response := testHandler(claims, config, request)
	expecthttp.Ok(t, response)

	expectedKey := &apikey.Struct{
		ID:        newAPIKey.ID,
		UserID:    user.ID,
		Name:      "test key",
		LastUsed:  newAPIKey.LastUsed,
		CreatedAt: newAPIKey.CreatedAt}

	expecthttp.JSONBody(t, expectedKey, response.Body)

	_, err = config.Store.GetByKey(newAPIKey.ID)
	expect.Error(t, err)
	expect.True(t, sql.ErrNoRows == err)

	// Deleting non existent key
	response = testHandler(claims, config, request)
	expecthttp.Status(t, http.StatusNotFound, response)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	config, user, teardown := testConfig(t)
	defer teardown()

	claims := &auth.Claims{UserID: user.ID}
	requestJson, err := utils.JSONBody(map[string]string{"name": "test key"})
	request, err := http.NewRequest("POST", "/apikeys", requestJson)
	expect.Ok(t, err)

	response := testHandler(claims, config, request)
	expecthttp.Ok(t, response)

	keys, err := config.Store.List(user.ID)
	expect.Ok(t, err)
	expect.Equal(t, 1, len(keys))

	var responseBody apikey.New
	err = json.Unmarshal(response.Body.Bytes(), &responseBody)
	expect.Ok(t, err)

	key := keys[0]
	expect.Equal(t, key.ID, responseBody.ID)
	expect.Equal(t, user.ID, responseBody.UserID)
	expect.Equal(t, "test key", responseBody.Name)
	expect.True(t, responseBody.Key != "")
}

func TestCreateStoreError(t *testing.T) {
	t.Parallel()

	config, _, teardown := testConfig(t)
	defer teardown()

	// Bad User
	claims := &auth.Claims{UserID: uuid.New().String()}
	requestJson, err := utils.JSONBody(map[string]string{"name": "test key"})
	request, err := http.NewRequest("POST", "/apikeys", requestJson)
	expect.Ok(t, err)

	response := testHandler(claims, config, request)
	expecthttp.Status(t, http.StatusUnprocessableEntity, response)

	// Empty body
	request, err = http.NewRequest("POST", "/apikeys", nil)
	expect.Ok(t, err)

	response = testHandler(claims, config, request)
	expecthttp.Status(t, http.StatusUnprocessableEntity, response)
}
