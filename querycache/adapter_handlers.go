package querycache

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cga1123/bissy-api/handlerutils"
)

func (c *Config) adaptersList(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	params := handlerutils.Params(r)
	page := params.MaybeInt("page", 1)
	per := params.MaybeInt("per", 25)

	adapters, err := c.AdapterStore.List(page, per)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}

	return json.NewEncoder(w).Encode(adapters)
}

func (c *Config) adaptersCreate(w http.ResponseWriter, r *http.Request) error {
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

	var createAdapter CreateAdapter
	if err := json.Unmarshal(body, &createAdapter); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	adapter, err := c.AdapterStore.Create(&createAdapter)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	err = json.NewEncoder(w).Encode(adapter)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}

	return nil
}

func (c *Config) adapterGet(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	params := handlerutils.Params(r)
	id, ok := params.Get("id")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
	}

	adapter, err := c.AdapterStore.Get(id)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(adapter)
}

func (c *Config) adapterDelete(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	params := handlerutils.Params(r)
	id, ok := params.Get("id")
	if !ok {
		return &handlerutils.HandlerError{
			Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
	}

	adapter, err := c.AdapterStore.Delete(id)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(adapter)
}

func (c *Config) adapterUpdate(w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	params := handlerutils.Params(r)
	id, ok := params.Get("id")
	if !ok {
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

	var updateAdapter UpdateAdapter
	if err := json.Unmarshal(body, &updateAdapter); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	adapter, err := c.AdapterStore.Update(id, &updateAdapter)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(adapter)
}
