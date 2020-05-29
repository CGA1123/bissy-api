// trevor -> rovert -> robert
package robert

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
	Store QueryStore
}

func (c *Config) SetupHandlers(router *mux.Router) {
	router.HandleFunc("/", c.Home).Methods("GET")
	router.Handle("/queries", &handlerutils.Handler{H: c.queriesCreate}).Methods("POST")
}

func (c *Config) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "robert, a poor man's trevor\nrobert -> trebor -> trevor\n")
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
