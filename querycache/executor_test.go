package querycache_test

import (
	"os"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/querycache"
	_ "github.com/lib/pq"
)

func TestExecutePostgres(t *testing.T) {
	t.Parallel()

	url, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		t.Fatal("DATABASE_URL not set")
	}

	executor, err := querycache.NewSQLExecutor("postgres", url)
	expect.Ok(t, err)

	query := &querycache.Query{Query: "SELECT * FROM (SELECT 1 a, 2 b) t"}
	csv, err := executor.Execute(query)
	expect.Ok(t, err)
	expect.Equal(t, "a,b\n1,2\n", csv)
}

func TestCachedExecutorExecute(t *testing.T) {
	t.Parallel()

	now := time.Now()

	executor := testCachedExecutor()
	query := &querycache.Query{
		ID:          "1",
		LastRefresh: now,
		UpdatedAt:   now.Add(-time.Second),
		Lifetime:    querycache.Duration(time.Hour),
		Query:       "SELECT 1;"}

	result, err := executor.Execute(query)
	expect.Ok(t, err)
	expect.Equal(t, "Got: SELECT 1;", result)

	query.Query = "SELECT 2;"
	result, err = executor.Execute(query)
	expect.Ok(t, err)
	expect.Equal(t, "Got: SELECT 1;", result)

	// When updated after last refresh
	query.UpdatedAt = query.LastRefresh.Add(time.Second)
	query.Query = "SELECT 3;"
	result, err = executor.Execute(query)
	expect.Ok(t, err)
	expect.Equal(t, "Got: SELECT 3;", result)

	// When LastRefresh longer than Lifetime ago
	result, err = executor.Execute(query)
	expect.Ok(t, err)
	expect.Equal(t, "Got: SELECT 3;", result)

	query.LastRefresh = now.Add(-time.Duration(query.Lifetime)).Add(-time.Second)
	query.Query = "SELECT 4;"
	result, err = executor.Execute(query)
	expect.Ok(t, err)
	expect.Equal(t, "Got: SELECT 4;", result)
}
