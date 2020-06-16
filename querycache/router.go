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

func handler(next func(*auth.Claims, http.ResponseWriter, *http.Request) error) http.Handler {
	return &handlerutils.Handler{
		H: auth.BuildHandler(next),
	}
}

func memberHandler(next func(*auth.Claims, string, http.ResponseWriter, *http.Request) error) http.Handler {
	return &handlerutils.Handler{
		H: auth.BuildHandler(
			func(claims *auth.Claims, w http.ResponseWriter, r *http.Request) error {
				handlerutils.ContentType(w, handlerutils.ContentTypeJSON)

				params := handlerutils.Params(r)
				id, ok := params.Get("id")
				if !ok {
					return &handlerutils.HandlerError{
						Err: fmt.Errorf("id not set"), Status: http.StatusBadRequest}
				}

				return next(claims, id, w, r)
			})}
}

// SetupHandlers mounts the querycache handlers onto the given mux
func (c *Config) SetupHandlers(router *mux.Router) {
	router.HandleFunc("/", c.home).Methods("GET")

	// Queries
	router.
		Handle("/queries", handler(c.queriesList)).
		Methods("GET")

	router.
		Handle("/queries", handler(c.queriesCreate)).
		Methods("POST")

	router.
		Handle("/queries/{id}", memberHandler(c.queryGet)).
		Methods("GET")

	router.
		Handle("/queries/{id}", memberHandler(c.queryDelete)).
		Methods("DELETE")

	router.
		Handle("/queries/{id}", memberHandler(c.queryUpdate)).
		Methods("PATCH")

	router.
		Handle("/queries/{id}/result", memberHandler(c.queryResult)).
		Methods("GET")

	// Datasources
	router.
		Handle("/datasources", handler(c.datasourcesList)).
		Methods("GET")

	router.
		Handle("/datasources", handler(c.datasourcesCreate)).
		Methods("POST")

	router.
		Handle("/datasources/{id}", memberHandler(c.datasourceGet)).
		Methods("GET")

	router.
		Handle("/datasources/{id}", memberHandler(c.datasourceDelete)).
		Methods("DELETE")

	router.
		Handle("/datasources/{id}", memberHandler(c.datasourceUpdate)).
		Methods("PATCH")
}

func (c *Config) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "querycache: using cache, saving cash\n")
}
