package utils_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/utils"
	_ "github.com/lib/pq"
)

func init() {
	url, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL not set")
	}

	txdb.Register("pgx", "postgres", url)
}

func TestHTTPClient(t *testing.T) {
	t.Parallel()

	client := utils.NewTestHTTPClient()

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
		return nil, fmt.Errorf("an error occurred")
	})

	response, err := client.Do(request)
	expect.Error(t, err)

	var nilResponse *http.Response

	expect.Equal(t, nilResponse, response)
	expect.Equal(t, "an error occurred", err.Error())

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

	testGenerator := &utils.TestIDGenerator{ID: "my-id"}
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

func TestTestDB(t *testing.T) {
	t.Parallel()

	db, teardown := utils.TestDB(t)
	defer teardown()

	err := db.Ping()
	expect.Ok(t, err)
}

func TestTestRedis(t *testing.T) {
	t.Parallel()

	redis, teardown := utils.TestRedis(t)
	defer teardown()

	err := redis.Ping(context.TODO()).Err()
	expect.Ok(t, err)
}

func TestJSONBody(t *testing.T) {
	createdAt, err := time.Parse(time.RFC3339, "1997-05-04T12:00:05+01:00")
	expect.Ok(t, err)

	json := struct {
		ID        string `json:"id"`
		Name      string
		CreatedAt time.Time `json:"createdAt"`
	}{
		ID:        "my=id",
		Name:      "Christian",
		CreatedAt: createdAt,
	}

	reader, err := utils.JSONBody(json)
	expect.Ok(t, err)

	var target struct {
		ID        string `json:"id"`
		Name      string
		CreatedAt time.Time `json:"createdAt"`
	}

	err = utils.ParseJSONBody(ioutil.NopCloser(reader), &target)
	expect.Ok(t, err)
	expect.Equal(t, json, target)

	// when passed junk
	_, err = utils.JSONBody(math.Inf(1))
	expect.Error(t, err)
}

func TestTestRandom(t *testing.T) {
	random := &utils.TestRandom{Value: []byte("hello, world")}

	str, err := random.String(0)
	expect.Ok(t, err)
	expect.Equal(t, "hello, world", str)

	bytes, err := random.Bytes(0)
	expect.Ok(t, err)
	expect.Equal(t, []byte("hello, world"), bytes)

	srandom := &utils.SecureRandom{}
	str1, err := srandom.String(32)
	expect.Ok(t, err)

	str2, err := srandom.String(32)
	expect.Ok(t, err)

	expect.NotEqual(t, str1, str2)
}
