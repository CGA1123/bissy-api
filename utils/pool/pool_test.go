package pool_test

import (
	"testing"

	"github.com/cga1123/bissy-api/utils/expect"
	"github.com/cga1123/bissy-api/utils/pool"
)

func TestPoolConformsToInterface(t *testing.T) {
	t.Parallel()

	var _ pool.Pool = pool.New(1, 1)
}

func TestPoolWork(t *testing.T) {
	x := 0

	p := pool.New(1, 10)

	for i := 0; i < 10; i++ {
		expect.True(t, p.Add(func() { x++ }))
	}

	expect.False(t, p.Add(func() {}))

	p.Start()
	p.Stop()

	expect.Equal(t, 10, x)
}
