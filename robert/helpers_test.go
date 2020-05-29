package robert_test

import (
	"fmt"
	"time"

	"github.com/cga1123/bissy-api/robert"
)

type testClock struct {
	time time.Time
}

func (clock *testClock) Now() time.Time {
	return clock.time
}

type testIdGenerator struct {
	id string
}

func (generator *testIdGenerator) Generate() string {
	return generator.id
}

func newTestStore(now time.Time, id string) *robert.InMemoryStore {
	return robert.NewInMemoryStore(
		&testClock{time: now},
		&testIdGenerator{id: id})
}

type testExecutor struct{}

func (t *testExecutor) Execute(query *robert.Query) (string, error) {
	return fmt.Sprintf("Got: %v", query.Query), nil
}
