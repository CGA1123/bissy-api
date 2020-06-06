package utils

import (
	"context"
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
	url, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		t.Fatal("REDIS_URL not set")
	}

	options, err := redis.ParseURL(url)
	expect.Ok(t, err)

	client := redis.NewClient(options)

	expect.Ok(t, client.Ping(context.TODO()).Err())

	return client, client.FlushAll(context.TODO()).Err
}
