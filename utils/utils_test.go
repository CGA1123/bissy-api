package utils_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

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

	request, err = http.NewRequest("GET", "https://example.com/notmocked", nil)
	expect.Ok(t, err)

	_, err = client.Do(request)
	expect.Error(t, err)
	expect.Equal(t, "no mock set for GET|https://example.com/notmocked", err.Error())
}

func TestRequireEnv(t *testing.T) {
	t.Parallel()

	os.Setenv("TEST_REQUIRE_ENV_1", "hello")
	os.Setenv("TEST_REQUIRE_ENV_2", "world")
	expected := map[string]string{
		"TEST_REQUIRE_ENV_1": "hello",
		"TEST_REQUIRE_ENV_2": "world",
	}

	vars, err := utils.RequireEnv("TEST_REQUIRE_ENV_1", "TEST_REQUIRE_ENV_2")
	expect.Ok(t, err)
	expect.Equal(t, expected, vars)

	_, err = utils.RequireEnv("TEST_REQUIRE_ENV_1", "NOPE")
	expect.Error(t, err)
	expect.Equal(t, "NOPE not set", err.Error())
}

func TestParseJSONBody(t *testing.T) {
	t.Parallel()

	var result struct {
		Hello      string `json:"hello"`
		AnotherKey string `json:"another_key"`
	}

	body := ioutil.NopCloser(bytes.NewReader([]byte(`
	{
		"hello": "what's up",
		"another_key": "yep"
	}`)))

	err := utils.ParseJSONBody(body, &result)
	expect.Ok(t, err)
	expect.Equal(t, result.Hello, "what's up")
	expect.Equal(t, result.AnotherKey, "yep")
}

func TestGenerators(t *testing.T) {
	t.Parallel()

	testGenerator := &utils.TestIdGenerator{Id: "my-id"}
	expect.Equal(t, testGenerator.Generate(), testGenerator.Generate())
	expect.Equal(t, testGenerator.Generate(), "my-id")

	uuidGenerator := &utils.UUIDGenerator{}
	expect.NotEqual(t, uuidGenerator.Generate(), uuidGenerator.Generate())
}

func TestClocks(t *testing.T) {
	testTime := time.Now().Add(-10 * time.Hour)
	testClock := &utils.TestClock{Time: testTime}

	expect.Equal(t, testTime, testClock.Now())

	clock := &utils.RealClock{}
	expect.Equal(t, clock.Now().Truncate(time.Millisecond), time.Now().Truncate(time.Millisecond))
}
