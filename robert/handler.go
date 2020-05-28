// trevor -> rovert -> robert
package robert

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

type Config struct {
	Store QueryStore
}

func (c *Config) SetupHandlers(router *mux.Router) {
	router.HandleFunc("", c.Home).Methods("GET")
	router.HandleFunc("/queries", c.IndexQueries).Methods("GET")
	router.HandleFunc("/queries", c.CreateQuery).Methods("POST")

	router.HandleFunc("/queries/{id}", c.ReadQuery).Methods("GET")
	router.HandleFunc("/queries/{id}", c.DeleteQuery).Methods("DELETE")
	router.HandleFunc("/queries/{id}", c.UpdateQuery).Methods("PATCH")
}

func (c *Config) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "robert, a poor man's trevor\nrobert -> trebor -> trevor\n")
}

// TODO:
// - Implement Index
// - Tests
// - Specs
// - Implement an adapter
// - Implement caching layer
// - Allow multi-adapter (add http adapter?)
//		- PG
//		- MySQL
// 		- Snowflake
//		- HTTP
// - Ensure Read-Only?

// TODO: Clean up error handling with func(http.ResponseWriter, *http.Request) error
//
// type HandlerError interface {
//		error
// 		Status()
// }
func intWithDefault(v url.Values, k string, d int) int {
	value, err := strconv.Atoi(v.Get(k))
	if err != nil {
		value = d
	}

	return value
}

func (c *Config) IndexQueries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	params := r.URL.Query()
	page := intWithDefault(params, "page", 1)
	per := intWithDefault(params, "per", 25)

	queries, err := c.Store.List(page, per)
	if err != nil {
		panic(err)
	}

	if err := json.NewEncoder(w).Encode(queries); err != nil {
		panic(err)
	}
}

func (c *Config) CreateQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	var createQuery CreateQuery
	if err := json.Unmarshal(body, &createQuery); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	query, err := c.Store.Create(&createQuery)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}

		return
	}

	json.NewEncoder(w).Encode(query)
}

func (c *Config) DeleteQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := mux.Vars(r)["id"]
	query, err := c.Store.Delete(id)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	json.NewEncoder(w).Encode(query)
}

func (c *Config) ReadQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := mux.Vars(r)["id"]
	query, err := c.Store.Get(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(query)
}

func (c *Config) UpdateQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := mux.Vars(r)["id"]

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	var updateQuery UpdateQuery
	if err := json.Unmarshal(body, &updateQuery); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	query, err := c.Store.Update(id, &updateQuery)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}

		return
	}

	json.NewEncoder(w).Encode(query)
}
