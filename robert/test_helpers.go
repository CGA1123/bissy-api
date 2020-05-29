package robert

import "time"

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

func newTestStore(now time.Time, id string) *InMemoryStore {
	return NewInMemoryStore(
		&testClock{time: now},
		&testIdGenerator{id: id})
}
