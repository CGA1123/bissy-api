package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Clock interface {
	Now() time.Time
}

type IdGenerator interface {
	Generate() string
}

type UUIDGenerator struct{}

func (generator *UUIDGenerator) Generate() string {
	return uuid.New().String()
}

type RealClock struct{}

func (clock *RealClock) Now() time.Time {
	return time.Now()
}

type TestClock struct {
	Time time.Time
}

func (clock *TestClock) Now() time.Time {
	return clock.Time
}

type TestIdGenerator struct {
	Id string
}

func (generator *TestIdGenerator) Generate() string {
	return generator.Id
}

func TestDB(t *testing.T) (*sqlx.DB, func() error) {
	db, err := sqlx.Open("pgx", uuid.New().String())
	expect.Ok(t, err)

	err = db.Ping()
	expect.Ok(t, err)

	return db, db.Close
}

func TestRedis(t *testing.T) (*redis.Client, func() error) {
	url := os.Getenv("REDISCLOUD_URL")
	options, err := redis.ParseURL(url)
	expect.Ok(t, err)

	client := redis.NewClient(options)

	expect.Ok(t, client.Ping(context.TODO()).Err())

	return client, client.FlushAll(context.TODO()).Err
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type requestHandler = func(*http.Request) (*http.Response, error)

type testHTTPRequest struct {
	method string
	url    string
}

type testHTTPMock struct {
	r *http.Response
	e error
}

type TestHTTPClient struct {
	Mocks map[string]requestHandler
	T     *testing.T
}

func NewTestHTTPClient(t *testing.T) *TestHTTPClient {
	return &TestHTTPClient{
		T:     t,
		Mocks: map[string]requestHandler{},
	}
}

func (c *TestHTTPClient) Do(r *http.Request) (*http.Response, error) {
	requestKey := r.Method + "|" + r.URL.String()
	responseFunc, ok := c.Mocks[requestKey]
	if !ok {
		return nil, fmt.Errorf("no mock set for %v", requestKey)
	}

	return responseFunc(r)
}

func (c *TestHTTPClient) Mock(method, url string, handler requestHandler) {
	c.Mocks[method+"|"+url] = handler
}

func JSONBody(v interface{}) (*bytes.Reader, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(body), nil
}

func ParseJSONBody(body io.ReadCloser, target interface{}) error {
	defer body.Close()

	return json.NewDecoder(body).Decode(target)
}

func RequireEnv(envVars ...string) (map[string]string, error) {
	vars := map[string]string{}

	for _, variableName := range envVars {
		value, ok := os.LookupEnv(variableName)
		if !ok {
			return map[string]string{}, fmt.Errorf("%v not set", variableName)
		}

		vars[variableName] = value
	}

	return vars, nil
}
