package handlerutils

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Defines some common ContentType strings
const (
	ContentTypeJSON      = "application/json; charset=UTF-8"
	ContentTypePlaintext = "text/plain"
	ContentTypeCSV       = "text/csv"
)

type handlerError interface {
	error
	StatusCode() int
}

// Handler extends the http.Handler interface and allows handlers to return and
// error
type Handler struct {
	H func(w http.ResponseWriter, r *http.Request) error
}

// HandlerError satisfies the error interface and allows for http handlers to
// set a specific HTTP Status Code
type HandlerError struct {
	Err    error
	Status int
}

// Error satisfies the error interface, returning the wrapped error
func (h *HandlerError) Error() string {
	return h.Err.Error()
}

// StatusCode returns the configured status code
func (h *HandlerError) StatusCode() int {
	return h.Status
}

// ContentType is a convience function to set the Content-Type header
func ContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

// ServeHTTP satifies the http.Handler interface and wraps handlers to deal with
// a returned error.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(w, r)
	if err != nil {
		log.Println("error:", err)
		switch e := err.(type) {
		case handlerError:
			http.Error(w, e.Error(), e.StatusCode())

		default:
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}
	}
}

// RequestParams represents the query and path parameters of a request
type RequestParams struct {
	values map[string][]string
}

// Params is a helper function that combines URL and query parameters into a
// single map
func Params(r *http.Request) *RequestParams {
	pathVars := mux.Vars(r)
	queryVars := r.URL.Query()
	params := map[string][]string{}

	for k, v := range queryVars {
		params[k] = v
	}

	for k, v := range pathVars {
		params[k] = []string{v}
	}

	return &RequestParams{values: params}
}

// Get returns the requested param, if available
func (p *RequestParams) Get(k string) (string, bool) {
	values, ok := p.values[k]
	if !ok || len(values) == 0 {
		return "", false
	}

	return values[0], true
}

// Int returns the given parameter casted to int if possible
func (p *RequestParams) Int(k string) (int, bool) {
	value, ok := p.Get(k)
	if !ok {
		return -1, false
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return -1, false
	}

	return i, true
}

// MaybeInt returns the given parameter casted to int, or the given fallback if
// not set or not castable
func (p *RequestParams) MaybeInt(k string, fallback int) int {
	value, ok := p.Int(k)
	if !ok {
		return fallback
	}

	return value
}
