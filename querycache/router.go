package querycache

import (
	"fmt"
	"net/http"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/handlerutils"
	"github.com/cga1123/bissy-api/utils"
	"github.com/gorilla/mux"
)

// Config contains everything external required to setup querycache
type Config struct {
	QueryStore      QueryStore
	DatasourceStore DatasourceStore
	Executor        Executor
	Cache           QueryCache
	Clock           utils.Clock
}

func memberHandler(next func(*auth.Claims, string, http.ResponseWriter, *http.Request) error) http.Handler {
	return auth.BuildHandler(
		func(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
			handlerutils.ContentType(w, handlerutils.ContentTypeJSON)

			params := handlerutils.Params(r)
			id, ok := params.Get("id")
			if !ok {
				return &handlerutils.HandlerError{
					Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
			}

			return next(claims, id, w, r)
		})
}

// SetupHandlers mounts the querycache handlers onto the given mux
func (c *Config) SetupHandlers(router *mux.Router) {
	router.HandleFunc("/", c.home).Methods("OPTIONS", "GET")

	// Queries
	router.
		Handle("/queries", auth.BuildHandler(c.queriesList)).
		Methods("OPTIONS", "GET")

	router.
		Handle("/queries", auth.BuildHandler(c.queriesCreate)).
		Methods("OPTIONS", "POST")

	router.
		Handle("/queries/{id}", memberHandler(c.queryGet)).
		Methods("OPTIONS", "GET")

	router.
		Handle("/queries/{id}", memberHandler(c.queryDelete)).
		Methods("OPTIONS", "DELETE")

	router.
		Handle("/queries/{id}", memberHandler(c.queryUpdate)).
		Methods("OPTIONS", "PATCH")

	router.
		Handle("/queries/{id}/result", memberHandler(c.queryResult)).
		Methods("OPTIONS", "GET")

	// Datasources
	router.
		Handle("/datasources", auth.BuildHandler(c.datasourcesList)).
		Methods("OPTIONS", "GET")

	router.
		Handle("/datasources", auth.BuildHandler(c.datasourcesCreate)).
		Methods("OPTIONS", "POST")

	router.
		Handle("/datasources/{id}", memberHandler(c.datasourceGet)).
		Methods("OPTIONS", "GET")

	router.
		Handle("/datasources/{id}", memberHandler(c.datasourceDelete)).
		Methods("OPTIONS", "DELETE")

	router.
		Handle("/datasources/{id}", memberHandler(c.datasourceUpdate)).
		Methods("OPTIONS", "PATCH")
}

func (c *Config) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "querycache: using cache, saving cash\n")
}
