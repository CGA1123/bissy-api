package utils_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/utils"
)

func TestHTTPClient(t *testing.T) {
	client := utils.NewTestHTTPClient(t)

	request, err := http.NewRequest("GET", "https://example.com/hello", nil)
	expect.Ok(t, err)

	client.Mock("GET", "https://example.com/hello", func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusFound,
		}, nil
	})

	res, err := client.Do(request)
	expect.Ok(t, err)
	expect.Equal(t, http.StatusFound, res.StatusCode)

	request, err = http.NewRequest("GET", "https://example.com/error", nil)
	expect.Ok(t, err)

	client.Mock("GET", "https://example.com/error", func(r *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("an error occured")
	})

	response, err := client.Do(request)
	expect.Error(t, err)

	var nilResponse *http.Response

	expect.Equal(t, nilResponse, response)
	expect.Equal(t, "an error occured", err.Error())
}
