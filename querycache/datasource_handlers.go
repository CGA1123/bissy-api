package querycache

import (
	"encoding/json"
	"net/http"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
)

func (c *Config) datasourcesList(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJSON)

	params := handlerutils.Params(r)
	page := params.MaybeInt("page", 1)
	per := params.MaybeInt("per", 25)

	datasources, err := c.DatasourceStore.List(claims.UserID, page, per)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusInternalServerError}
	}

	return json.NewEncoder(w).Encode(datasources)
}

func (c *Config) datasourcesCreate(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
	handlerutils.ContentType(w, handlerutils.ContentTypeJSON)

	var createDatasource CreateDatasource
	if err := utils.ParseJSONBody(r.Body, &createDatasource); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	datasource, err := c.DatasourceStore.Create(claims.UserID, &createDatasource)
	if err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	return json.NewEncoder(w).Encode(datasource)
}

func (c *Config) datasourceGet(claims *auth.Claims, id string, w http.ResponseWriter, r *http.Request) error {
	datasource, err := c.DatasourceStore.Get(claims.UserID, id)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(datasource)
}

func (c *Config) datasourceDelete(claims *auth.Claims, id string, w http.ResponseWriter, r *http.Request) error {
	datasource, err := c.DatasourceStore.Delete(claims.UserID, id)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(datasource)
}

func (c *Config) datasourceUpdate(claims *auth.Claims, id string, w http.ResponseWriter, r *http.Request) error {
	var updateDatasource UpdateDatasource
	if err := utils.ParseJSONBody(r.Body, &updateDatasource); err != nil {
		return &handlerutils.HandlerError{
			Err: err, Status: http.StatusUnprocessableEntity}
	}

	datasource, err := c.DatasourceStore.Update(claims.UserID, id, &updateDatasource)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(datasource)
}
