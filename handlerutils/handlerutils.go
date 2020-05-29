package handlerutils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	ContentTypeJson      = "application/json; charset=UTF-8"
	ContentTypePlaintext = "text/plain"
	ContentTypeCsv       = "text/csv"
)

type requestParams struct {
	Values map[string][]string
}

type handlerError interface {
	error
	StatusCode() int
}

type Handler struct {
	H func(w http.ResponseWriter, r *http.Request) error
}

type HandlerError struct {
	Err    error
	Status int
}

func (h *HandlerError) Error() string {
	return h.Err.Error()
}

func (h *HandlerError) StatusCode() int {
	return h.Status
}

func ContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(w, r)
	if err != nil {
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

func Params(r *http.Request) *requestParams {
	pathVars := mux.Vars(r)
	queryVars := r.URL.Query()
	params := map[string][]string{}

	for k, v := range queryVars {
		params[k] = v
	}

	for k, v := range pathVars {
		params[k] = []string{v}
	}

	return &requestParams{Values: params}
}

func (p *requestParams) Get(k string) (string, error) {
	values, ok := p.Values[k]
	if !ok || len(values) == 0 {
		return "", fmt.Errorf("no value for: %v", k)
	}

	return values[0], nil
}

func (p *requestParams) Int(k string) (int, error) {
	value, err := p.Get(k)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(value)
}

func (p *requestParams) MaybeInt(k string, fallback int) int {
	value, err := p.Int(k)
	if err != nil {
		return fallback
	}

	return value
}
