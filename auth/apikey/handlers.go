package apikey

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/utils"
	"github.com/cga1123/bissy-api/utils/handlerutils"
	"github.com/gorilla/mux"
)

// Config holds the configuration for serving the apikey endpoints
type Config struct {
	Store Store
}

// SetupHandlers adds the apikey HTTP handlers to the given router
func (c *Config) SetupHandlers(router *mux.Router) {
	router.
		Handle("/apikeys", auth.BuildHandler(c.apikeysList)).
		Methods("OPTIONS", "GET")

	router.
		Handle("/apikeys", auth.BuildHandler(c.apikeysCreate)).
		Methods("OPTIONS", "POST")

	router.
		Handle("/apikeys/{id}", auth.BuildHandler(c.apikeysDelete)).
		Methods("OPTIONS", "DELETE")
}

func (c *Config) apikeysList(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJSON)

	keys, err := c.Store.List(claims.UserID)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}

	return json.NewEncoder(w).Encode(keys)
}

func (c *Config) apikeysDelete(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJSON)

	params := handlerutils.Params(r)
	id, ok := params.Get("id")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
	}

	key, err := c.Store.Delete(claims.UserID, id)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(key)
}

func (c *Config) apikeysCreate(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJSON)

	var create Create
	if err := utils.ParseJSONBody(r.Body, &create); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	key, err := c.Store.Create(claims.UserID, &create)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	return json.NewEncoder(w).Encode(key)
}
