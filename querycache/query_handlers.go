package querycache

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
)

func (c *Config) queriesList(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJSON)

	params := handlerutils.Params(r)
	page := params.MaybeInt("page", 1)
	per := params.MaybeInt("per", 25)

	queries, err := c.QueryStore.List(page, per)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}

	return json.NewEncoder(w).Encode(queries)
}

func (c *Config) queriesCreate(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJSON)

	var createQuery CreateQuery
	if err := utils.ParseJSONBody(r.Body, &createQuery); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	query, err := c.QueryStore.Create(claims.UserID, &createQuery)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	return json.NewEncoder(w).Encode(query)
}

func (c *Config) queryGet(claims *auth.Claims, id string, w http.ResponseWriter, r *http.Request) error {
	query, err := c.QueryStore.Get(claims.UserID, id)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(query)
}

func (c *Config) queryDelete(claims *auth.Claims, id string, w http.ResponseWriter, r *http.Request) error {
	query, err := c.QueryStore.Delete(id)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(query)
}

func (c *Config) queryUpdate(claims *auth.Claims, id string, w http.ResponseWriter, r *http.Request) error {
	var updateQuery UpdateQuery
	if err := utils.ParseJSONBody(r.Body, &updateQuery); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	query, err := c.QueryStore.Update(id, &updateQuery)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(query)
}

func (c *Config) queryResult(claims *auth.Claims, id string, w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeCSV)

	query, err := c.QueryStore.Get(claims.UserID, id)
	if err != nil {
		return err
	}

	// TODO: no need to pass claims one Query has a UserID!
	result, err := c.executeQuery(claims.UserID, query)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, result)

	return err
}

func (c *Config) executeQuery(userID string, query *Query) (string, error) {
	datasource, err := c.DatasourceStore.Get(userID, query.DatasourceID)
	if err != nil {
		return "", err
	}

	executor, err := datasource.NewExecutor()
	if err != nil {
		return "", err
	}

	if c.Cache != nil {
		executor = NewCachedExecutor(c.Cache, c.QueryStore, c.Clock, executor)
	}

	return executor.Execute(query)
}
