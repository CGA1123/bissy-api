package robert

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/snowflakedb/gosnowflake"
)

type Executor interface {
	Execute(string) (string, error)
}

type SQLExecutor struct {
	db *sql.DB
}

func NewSQLExecutor(driver, conn string) (*SQLExecutor, error) {
	db, err := sql.Open(driver, conn)
	if err != nil {
		return nil, err
	}

	return &SQLExecutor{db: db}, db.Ping()
}

func (sql *SQLExecutor) Execute(query string) (string, error) {
	rows, err := sql.db.Query(query)
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
