package querycache

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"sync"
	"time"

	"github.com/cga1123/bissy-api/utils"
	"github.com/go-redis/redis/v8"

	// add support for mysql databases
	_ "github.com/go-sql-driver/mysql"
	// add support for postgres databases
	_ "github.com/lib/pq"
	// add support for snowflake databases
	_ "github.com/snowflakedb/gosnowflake"
)

// QueryCache defines the interface for a cache of query results
type QueryCache interface {
	Get(*Query) (string, bool)
	Set(*Query, string) error
}

// InMemoryCache is an in-memory implementation of QueryCache
type InMemoryCache struct {
	Cache map[string]string
	lock  sync.RWMutex
}

// NewInMemoryCache sets up a new InMemoryCache
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{Cache: map[string]string{}}
}

// Get returns the cached results for a given query
func (cache *InMemoryCache) Get(query *Query) (string, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	value, ok := cache.Cache[query.ID]
	return value, ok
}

// Set caches the results for a given query
func (cache *InMemoryCache) Set(query *Query, result string) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	cache.Cache[query.ID] = result

	return nil
}

// RedisCache is a redis-backed implementation of QueryCache
type RedisCache struct {
	Client *redis.Client
}

// Get returns the cached results for a given query
func (cache *RedisCache) Get(query *Query) (string, bool) {
	value, err := cache.Client.Get(
		context.TODO(),
		"querycache:"+query.ID,
	).Result()

	return value, err == nil
}

// Set caches the results for a given query
func (cache *RedisCache) Set(query *Query, result string) error {
	set := cache.Client.Set(
		context.TODO(),
		"querycache:"+query.ID,
		result,
		time.Duration(query.Lifetime))

	return set.Err()
}

// Executor defines the interface to execute a query
type Executor interface {
	Execute(*Query) (string, error)
}

// TestExecutor implements an Executor that echoes the passed query
type TestExecutor struct{}

// Execute echoes the given query
func (t *TestExecutor) Execute(query *Query) (string, error) {
	return fmt.Sprintf("Got: %v", query.Query), nil
}

// CachedExecutor implements Executor that caches query results for the given
// Lifetime of a Query
type CachedExecutor struct {
	Cache    QueryCache
	Executor Executor
	Store    QueryStore
	Clock    utils.Clock
}

func updateCache(cache *CachedExecutor, query *Query, result string) {
	if err := cache.Cache.Set(query, result); err != nil {
		return
	}

	cache.Store.Update(query.ID, &UpdateQuery{LastRefresh: cache.Clock.Now()})
}

// NewCachedExecutor sets up a new CachedExecutor
func NewCachedExecutor(cache QueryCache, store QueryStore, clock utils.Clock, executor Executor) *CachedExecutor {
	return &CachedExecutor{
		Cache:    cache,
		Store:    store,
		Clock:    clock,
		Executor: executor,
	}
}

// Execute checks the cache for the given query cache, fallsback to the the
// configured executor if no results are found and stores the new results.
func (cache *CachedExecutor) Execute(query *Query) (string, error) {
	if query.Fresh(cache.Clock.Now()) {
		if result, ok := cache.Cache.Get(query); ok {
			return result, nil
		}
	}

	result, err := cache.Executor.Execute(query)
	if err != nil {
		return "", err
	}

	updateCache(cache, query, result)

	return result, nil
}

// SQLExecutor implements Executor against an *sql.DB
type SQLExecutor struct {
	db *sql.DB
}

// NewSQLExecutor builds a new SQLExecutor, parameters are passed to sql.Open
func NewSQLExecutor(driver, conn string) (*SQLExecutor, error) {
	db, err := sql.Open(driver, conn)
	if err != nil {
		return nil, err
	}

	return &SQLExecutor{db: db}, nil
}

// Execute runs the query against the configured database
// returns the results as a CSV string
func (sql *SQLExecutor) Execute(query *Query) (string, error) {
	rows, err := sql.db.Query(query.Query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	results, err := parseRows(rows)
	if err != nil {
		return "", err
	}

	return resultsToCSVString(*results)
}

func parseRows(rows *sql.Rows) (*[][]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := [][]string{cols}

	count := len(cols)
	vals := make([]interface{}, count)
	ptrs := make([]interface{}, count)

	for rows.Next() {
		row := make([]string, count)
		for i := range cols {
			ptrs[i] = &vals[i]
		}

		if err = rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		for i := range cols {
			row[i] = parseColumnValue(vals[i])
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &results, nil
}

func resultsToCSVString(rows [][]string) (string, error) {
	buffer := bytes.Buffer{}
	csvWriter := csv.NewWriter(&buffer)
	if err := csvWriter.WriteAll(rows); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func parseColumnValue(rawValue interface{}) string {
	var value interface{}

	switch v := rawValue.(type) {
	case []byte:
		value = string(v)
	default:
		value = v
	}

	if value == nil {
		return ""
	}

	return fmt.Sprintf("%v", value)
}
