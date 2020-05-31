package querycache

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cga1123/bissy-api/handlerutils"
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

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}
	if err := r.Body.Close(); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}

	var createQuery CreateQuery
	if err := json.Unmarshal(body, &createQuery); err != nil {
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
	id, err := params.Get("id")
	if err != nil {
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
	id, err := params.Get("id")
	if err != nil {
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
	id, err := params.Get("id")
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
	}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}
	if err := r.Body.Close(); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}

	var updateQuery UpdateQuery
	if err := json.Unmarshal(body, &updateQuery); err != nil {
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
	id, err := params.Get("id")
	if err != nil {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
	}

	query, err := c.QueryStore.Get(id)
	if err != nil {
		return err
	}

	adapter, err := c.AdapterStore.Get(query.AdapterId)
	if err != nil {
		return err
	}

	executor, err := adapter.NewExecutor()
	if err != nil {
		return err
	}

	result, err := executor.Execute(query)
	_, err = fmt.Fprintf(w, result)

	return err
}
