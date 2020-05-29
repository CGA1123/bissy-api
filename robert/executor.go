package robert

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"time"
)

type QueryCache interface {
	Get(*Query) (string, bool)
	Set(*Query, string) error
}

type Executor interface {
	Execute(*Query) (string, error)
}

type CachedExecutor struct {
	Cache    QueryCache
	Executor Executor
}

type SQLExecutor struct {
	db *sql.DB
}

func (cache *CachedExecutor) Execute(query *Query) (string, error) {
	if result, ok := cache.Cache.Get(query); ok {
		return result, nil
	}

	result, err := cache.Executor.Execute(query)
	if err != nil {
		return "", err
	}

	err = cache.Cache.Set(query, result)
	if err != nil {
		log.Printf("could not set cache: %v", err)
	}

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
