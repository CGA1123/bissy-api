package handlerutils_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cga1123/bissy-api/utils/expect"
	"github.com/cga1123/bissy-api/utils/expecthttp"
	"github.com/cga1123/bissy-api/utils/handlerutils"
	"github.com/gorilla/mux"
)

func TestRequire(t *testing.T) {
	t.Parallel()

	params := handlerutils.ParamsFromMap(map[string][]string{
		"another-key": []string{"goodbye"},
		"this-key":    []string{"hello"}})

	expect.Ok(t, params.Require("this-key"))
	expect.Ok(t, params.Require("this-key", "another-key"))
	expect.Error(t, params.Require("that-key"))
	expect.Error(t, params.Require("this-key", "another-key", "that-key"))
}

func TestHandlerError(t *testing.T) {
	t.Parallel()

	handlerError := &handlerutils.HandlerError{Err: fmt.Errorf("test"), Status: http.StatusBadRequest}

	// check we conform to error interface
	// tests will fail to compile if not.
	var _ error = handlerError

	expect.Equal(t, "test", handlerError.Error())
	expect.Equal(t, http.StatusBadRequest, handlerError.StatusCode())
}

func TestGet(t *testing.T) {
	t.Parallel()

	params := handlerutils.ParamsFromMap(map[string][]string{
		"another-key": []string{"goodbye"},
		"int-key":     []string{"1123"},
		"this-key":    []string{"hello"}})

	// Get
	thisValue, thisOk := params.Get("this-key")
	anotherValue, anotherOk := params.Get("another-key")
	_, thatOk := params.Get("that-key")

	expect.True(t, thisOk)
	expect.Equal(t, "hello", thisValue)

	expect.True(t, anotherOk)
	expect.Equal(t, "goodbye", anotherValue)

	expect.False(t, thatOk)

	// Int
	intValue, intOk := params.Int("int-key")
	_, notIntOk := params.Int("this-key")
	_, notExistOk := params.Int("that-keu")

	expect.True(t, intOk)
	expect.Equal(t, 1123, intValue)

	expect.False(t, notIntOk)
	expect.False(t, notExistOk)

	// MaybeInt
	isIntValue := params.MaybeInt("int-key", 3211)
	notIntValue := params.MaybeInt("this-key", 3211)
	notExistValue := params.MaybeInt("that-key", 3211)

	expect.Equal(t, 1123, isIntValue)
	expect.Equal(t, 3211, notIntValue)
	expect.Equal(t, 3211, notExistValue)
}

func TestParams(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	handler := func(w http.ResponseWriter, r *http.Request) {
		params := handlerutils.Params(r)

		pathParam, pathOk := params.Get("pathParam")
		queryParam, queryOk := params.Get("queryParam")

		expect.True(t, pathOk)
		expect.Equal(t, "path-parameter", pathParam)

		expect.True(t, queryOk)
		expect.Equal(t, "query-parameter", queryParam)

	}

	router.HandleFunc("/test/{pathParam}", handler).Methods("GET")

	request, err := http.NewRequest("GET", "/test/path-parameter?queryParam=query-parameter", nil)
	expect.Ok(t, err)

	router.ServeHTTP(recorder, request)
}

func TestHandler(t *testing.T) {
	t.Parallel()

	// no error
	handler := &handlerutils.Handler{
		H: func(w http.ResponseWriter, r *http.Request) error { return nil }}
	sqlErrHandler := &handlerutils.Handler{
		H: func(w http.ResponseWriter, r *http.Request) error { return sql.ErrNoRows }}
	handlerErrHandler := &handlerutils.Handler{H: func(w http.ResponseWriter, r *http.Request) error {
		return &handlerutils.HandlerError{Err: fmt.Errorf("error"), Status: http.StatusBadRequest}
	}}

	otherErrHandler := &handlerutils.Handler{H: func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf("error")
	}}

	router := mux.NewRouter()
	router.Handle("/handler", handler).Methods("GET")
	router.Handle("/sqlErrHandler", sqlErrHandler).Methods("GET")
	router.Handle("/handlerErrHandler", handlerErrHandler).Methods("GET")
	router.Handle("/otherErrHandler", otherErrHandler).Methods("GET")

	handlerRequest, err := http.NewRequest("GET", "/handler", nil)
	expect.Ok(t, err)

	sqlErrHandlerRequest, err := http.NewRequest("GET", "/sqlErrHandler", nil)
	expect.Ok(t, err)

	handlerErrHandlerRequest, err := http.NewRequest("GET", "/handlerErrHandler", nil)
	expect.Ok(t, err)

	otherErrHandlerRequest, err := http.NewRequest("GET", "/otherErrHandler", nil)
	expect.Ok(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, handlerRequest)
	expecthttp.Ok(t, recorder)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, sqlErrHandlerRequest)
	expecthttp.Status(t, http.StatusNotFound, recorder)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, handlerErrHandlerRequest)
	expecthttp.Status(t, http.StatusBadRequest, recorder)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, otherErrHandlerRequest)
	expecthttp.Status(t, http.StatusInternalServerError, recorder)
}

func TestContentType(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) { handlerutils.ContentType(w, "testing") }

	router := mux.NewRouter()
	router.HandleFunc("/test/{pathParam}", handler).Methods("GET")

	request, err := http.NewRequest("GET", "/test/path-parameter?queryParam=query-parameter", nil)
	expect.Ok(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	expecthttp.Header(t, "Content-Type", "testing", recorder.Header())
}
