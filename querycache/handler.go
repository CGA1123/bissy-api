package querycache

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/gorilla/mux"
)

type Config struct {
	Store    QueryStore
	Executor Executor
}

func (c *Config) SetupHandlers(router *mux.Router) {
	router.HandleFunc("/", c.Home).Methods("GET")

	router.
		Handle("/queries", &handlerutils.Handler{H: c.queriesList}).
		Methods("GET")

	router.
		Handle("/queries", &handlerutils.Handler{H: c.queriesCreate}).
		Methods("POST")

	router.
		Handle("/queries/{id}", &handlerutils.Handler{H: c.queryGet}).
		Methods("GET")

	router.
		Handle("/queries/{id}", &handlerutils.Handler{H: c.queryDelete}).
		Methods("DELETE")

	router.
		Handle("/queries/{id}", &handlerutils.Handler{H: c.queryUpdate}).
		Methods("PATCH")

	router.
		Handle("/queries/{id}/result", &handlerutils.Handler{H: c.queryResult}).
		Methods("GET")
}

func (c *Config) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "querycache: using cache, saving cash\n")
}

func (c *Config) queriesList(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	params := handlerutils.Params(r)
	page := params.MaybeInt("page", 1)
	per := params.MaybeInt("per", 25)

	queries, err := c.Store.List(page, per)
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

	query, err := c.Store.Create(&createQuery)
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

	query, err := c.Store.Get(id)
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

	query, err := c.Store.Delete(id)
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

	query, err := c.Store.Update(id, &updateQuery)
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

	query, err := c.Store.Get(id)
	if err != nil {
		return err
	}
	result, err := c.Executor.Execute(query)
	_, err = fmt.Fprintf(w, result)

	return err
}
