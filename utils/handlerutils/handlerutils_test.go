package handlerutils_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cga1123/bissy-api/utils/expect"
	"github.com/cga1123/bissy-api/utils/handlerutils"
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
