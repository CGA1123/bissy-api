package querycache

import (
	"fmt"
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
