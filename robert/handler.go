// trevor -> rovert -> robert
package robert

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Config struct{}

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

func (c *Config) IndexQueries(w http.ResponseWriter, r *http.Request) {}
func (c *Config) CreateQuery(w http.ResponseWriter, r *http.Request)  {}
func (c *Config) DeleteQuery(w http.ResponseWriter, r *http.Request)  {}
func (c *Config) ReadQuery(w http.ResponseWriter, r *http.Request)    {}
func (c *Config) UpdateQuery(w http.ResponseWriter, r *http.Request)  {}
