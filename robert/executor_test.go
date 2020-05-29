package robert

import (
	"testing"

	"github.com/cga1123/bissy-api/expect"
)

func TestExecutePostgres(t *testing.T) {
	t.Parallel()

	executor, err := NewSQLExecutor("postgres", "sslmode=disable")
	expect.Ok(t, err)

	csv, err := executor.Execute("SELECT * FROM (SELECT 1 a, 2 b) t")
	expect.Ok(t, err)
	expect.Equal(t, "a,b\n1,2\n", csv)

}
