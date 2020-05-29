package robert

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"time"
)

type QueryCache interface {
	Get(*Query) (string, bool)
	Set(*Query, string) error
}

type InMemoryCache struct {
	Cache map[string]string
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{Cache: map[string]string{}}
}
func (cache *InMemoryCache) Get(query *Query) (string, bool) {
	value, ok := cache.Cache[query.Id]
	return value, ok
}

func (cache *InMemoryCache) Set(query *Query, result string) error {
	cache.Cache[query.Id] = result

	return nil
}

type Executor interface {
	Execute(*Query) (string, error)
}

type CachedExecutor struct {
	Cache    QueryCache
	Executor Executor
	Store    QueryStore
	Clock    Clock
}

type SQLExecutor struct {
	db *sql.DB
}

func updateCache(cache *CachedExecutor, query *Query, result string) {
	if err := cache.Cache.Set(query, result); err != nil {
		return
	}

	cache.Store.Update(query.Id, &UpdateQuery{LastRefresh: time.Now()})
}

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

func NewSQLExecutor(driver, conn string) (*SQLExecutor, error) {
	db, err := sql.Open(driver, conn)
	if err != nil {
		return nil, err
	}

	return &SQLExecutor{db: db}, db.Ping()
}

func (sql *SQLExecutor) Execute(query *Query) (string, error) {
	rows, err := sql.db.Query(query.Query)
	if err != nil {
		return "", err
	}

	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	results := [][]string{cols}

	count := len(cols)
	vals := make([]interface{}, count)
	ptrs := make([]interface{}, count)
	for rows.Next() {
		row := make([]string, count)
		for i, _ := range cols {
			ptrs[i] = &vals[i]
		}

		if err = rows.Scan(ptrs...); err != nil {
			return "", err
		}

		for i, _ := range cols {
			var value interface{}
			rawValue := vals[i]

			byteArray, ok := rawValue.([]byte)
			if ok {
				value = string(byteArray)
			} else {
				value = rawValue
			}

			timeValue, ok := value.(time.Time)
			if ok {
				value = timeValue.Format(time.RFC3339)
			}

			if value == nil {
				row[i] = ""
			} else {
				row[i] = fmt.Sprintf("%v", value)
			}
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	buffer := bytes.Buffer{}
	csvWriter := csv.NewWriter(&buffer)
	if err := csvWriter.WriteAll(results); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
