package querycache

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
)

func (c *Config) queriesList(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

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

func (c *Config) queriesCreate(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	var createQuery CreateQuery
	if err := utils.ParseJSONBody(r.Body, &createQuery); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	query, err := c.QueryStore.Create(&createQuery)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	err = json.NewEncoder(w).Encode(query)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}

	return nil
}

func (c *Config) queryGet(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	params := handlerutils.Params(r)
	id, ok := params.Get("id")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
	}

	query, err := c.QueryStore.Get(id)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(query)
}

func (c *Config) queryDelete(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	params := handlerutils.Params(r)
	id, ok := params.Get("id")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
	}

	query, err := c.QueryStore.Delete(id)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(query)
}

func (c *Config) queryUpdate(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	params := handlerutils.Params(r)
	id, ok := params.Get("id")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
	}

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

func (c *Config) queryResult(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeCsv)

	params := handlerutils.Params(r)
	id, ok := params.Get("id")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
	}

	query, err := c.QueryStore.Get(id)
	if err != nil {
		return err
	}

	datasource, err := c.DatasourceStore.Get(query.DatasourceId)
	if err != nil {
		return err
	}

	executor, err := datasource.NewExecutor()
	if err != nil {
		return err
	}

	if c.Cache != nil {
		executor = NewCachedExecutor(c.Cache, c.QueryStore, c.Clock, executor)
	}

	result, err := executor.Execute(query)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, result)

	return err
}
