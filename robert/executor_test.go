package robert_test

import (
	"testing"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/robert"
	_ "github.com/lib/pq"
)

func TestExecutePostgres(t *testing.T) {
	t.Parallel()

	executor, err := robert.NewSQLExecutor("postgres", "sslmode=disable")
	expect.Ok(t, err)

	query := &robert.Query{Query: "SELECT * FROM (SELECT 1 a, 2 b) t"}
	csv, err := executor.Execute(query)
	expect.Ok(t, err)
	expect.Equal(t, "a,b\n1,2\n", csv)
}
