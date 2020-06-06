package querycache

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
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

	var createAdapter CreateAdapter
	if err := utils.ParseJSONBody(r.Body, &createAdapter); err != nil {
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

	var updateAdapter UpdateAdapter
	if err := utils.ParseJSONBody(r.Body, &updateAdapter); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	adapter, err := c.AdapterStore.Update(id, &updateAdapter)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(adapter)
}
