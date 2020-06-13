package auth_test

import (
	"testing"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/utils"
)

func TestRedisStateStore(t *testing.T) {
	redis, teardown := utils.TestRedis(t)
	defer teardown()

	store := &auth.RedisStateStore{Client: redis, IDGenerator: &utils.UUIDGenerator{}}

	key, err := store.Set("hello", time.Hour)
	expect.Ok(t, err)

	value, err := store.Get(key)
	expect.Ok(t, err)
	expect.Equal(t, "hello", value)

	exists, err := store.Exists(key)
	expect.Ok(t, err)
	expect.True(t, exists)

	value, err = store.Get(key)
	expect.Ok(t, err)
	expect.Equal(t, "hello", value)

	deleted, err := store.Del(key)
	expect.Ok(t, err)
	expect.True(t, deleted)

	_, err = store.Get(key)
	expect.Error(t, err)

	exists, err = store.Exists(key)
	expect.Ok(t, err)
	expect.False(t, exists)

	deleted, err = store.Del(key)
	expect.Ok(t, err)
	expect.False(t, deleted)
}
