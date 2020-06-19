package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
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
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
	"github.com/jmoiron/sqlx"
)

// The Clock interface implements Now(), useful in order to inject specific
// times during testing
type Clock interface {
	Now() time.Time
}

// IDGenerator interface is used to implement uuid generators
type IDGenerator interface {
	Generate() string
}

// UUIDGenerator implements IdGenerator via google/uuid generation
type UUIDGenerator struct{}

// Generate returns a random UUID string
func (generator *UUIDGenerator) Generate() string {
	return uuid.New().String()
}

// RealClock Implements Clock by delegating to time.Now()
type RealClock struct{}

// Now returns the current time
func (clock *RealClock) Now() time.Time {
	return time.Now()
}

// TestClock implements Clock with a constant value
type TestClock struct {
	Time time.Time
}

// Now returns the Time parameter of TestClock
func (clock *TestClock) Now() time.Time {
	return clock.Time
}

// TestIDGenerator implements IDGenerator with a constant value
type TestIDGenerator struct {
	ID string
}

// Generate returns the ID parameter of TestIDGenerator
func (generator *TestIDGenerator) Generate() string {
	return generator.ID
}

// TestDB returns a test db connection
func TestDB(t *testing.T) (*hnysqlx.DB, func() error) {
	db, err := sqlx.Open("pgx", uuid.New().String())
	expect.Ok(t, err)

	err = db.Ping()
	expect.Ok(t, err)

	return hnysqlx.WrapDB(db), db.Close
}

// TestRedis returns a test redis connection
func TestRedis(t *testing.T) (*redis.Client, func() error) {
	url := os.Getenv("REDISCLOUD_URL")
	options, err := redis.ParseURL(url)
	expect.Ok(t, err)

	client := redis.NewClient(options)

	expect.Ok(t, client.Ping(context.TODO()).Err())

	return client, client.FlushAll(context.TODO()).Err
}

// HTTPClient describes the interface for an HTTP client
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

// TestHTTPClient struct represents a mocked HTTPClient
type TestHTTPClient struct {
	mocks map[string]requestHandler
	t     *testing.T
}

// NewTestHTTPClient creates a new TestHTTPClient struct
func NewTestHTTPClient() *TestHTTPClient {
	return &TestHTTPClient{
		mocks: map[string]requestHandler{},
	}
}

// Do satisfies the HTTPClient interface, allowing the TestHTTPClient to be used
// in place of real HTTPClients in tests.
func (c *TestHTTPClient) Do(r *http.Request) (*http.Response, error) {
	requestKey := r.Method + "|" + r.URL.String()
	responseFunc, ok := c.mocks[requestKey]
	if !ok {
		return nil, fmt.Errorf("no mock set for %v", requestKey)
	}

	return responseFunc(r)
}

// Mock sets up a new mocked handler for the given method and URL
func (c *TestHTTPClient) Mock(method, url string, handler requestHandler) {
	c.mocks[method+"|"+url] = handler
}

// JSONBody returns the serialized JSON of the given value as a *bytes.Reader
// Useful for populating HTTP request bodies
func JSONBody(v interface{}) (*bytes.Reader, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(body), nil
}

// ParseJSONBody parse the given body into target using a JSON Decoder
func ParseJSONBody(body io.ReadCloser, target interface{}) error {
	defer body.Close()

	return json.NewDecoder(body).Decode(target)
}

// RequireEnv takes a variable number of environment variables and ensures they
// are all set
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

// Random describes an interface to get random strings and bytes
type Random interface {
	String(int) (string, error)
	Bytes(int) ([]byte, error)
}

// SecureRandom implements Random using the crypto/rand package
type SecureRandom struct{}

// Bytes returns a slice of random bytes of given size
func (s *SecureRandom) Bytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}

// String returns a Base-64 encoded string of random bytes of given size
func (s *SecureRandom) String(size int) (string, error) {
	bytes, err := s.Bytes(size)
	return base64.URLEncoding.EncodeToString(bytes), err
}

// TestRandom implements the Random implementation by returning a known value.
type TestRandom struct {
	Value []byte
}

// Bytes returns the configured Bytes
func (s *TestRandom) Bytes(size int) ([]byte, error) {
	return s.Value, nil
}

// String returns the configured Bytes as a String
func (s *TestRandom) String(size int) (string, error) {
	return string(s.Value), nil
}
