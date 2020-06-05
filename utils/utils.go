package utils

import (
	"time"

	"github.com/google/uuid"
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
