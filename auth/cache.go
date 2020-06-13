package auth

import (
	"context"
	"time"

	"github.com/cga1123/bissy-api/utils"
	"github.com/go-redis/redis/v8"
)

// StateStore defines an interface for caching state parameters during the
// O authentication processes
type StateStore interface {
	Exists(string) (bool, error)
	Set(string, time.Duration) (string, error)
	Del(string) (bool, error)
	Get(string) (string, error)
}

// RedisStateStore is a redis backed implementation of StateStore
type RedisStateStore struct {
	Client      *redis.Client
	IDGenerator utils.IDGenerator
}

// Get returns the payload stored at a key
func (r *RedisStateStore) Get(key string) (string, error) {
	value, err := r.Client.Get(context.TODO(), "auth:"+key).Result()
	if err != nil {
		return "", err
	}

	return value, nil
}

// Exists checks whether the given key exists in cache
func (r *RedisStateStore) Exists(key string) (bool, error) {
	value, err := r.Client.Exists(context.TODO(), "auth:"+key).Result()
	if err != nil {
		return false, err
	}

	return value == 1, nil
}

// Set stores the payload under a new random key, which is returned.
// exp sets the lifetime for the key.
func (r *RedisStateStore) Set(payload string, exp time.Duration) (string, error) {
	key := r.IDGenerator.Generate()
	_, err := r.Client.Set(
		context.TODO(),
		"auth:"+key,
		payload,
		exp,
	).Result()

	return key, err
}

// Del removes the given key
// returns true if a key was removed.
func (r *RedisStateStore) Del(key string) (bool, error) {
	count, err := r.Client.Del(context.TODO(), "auth:"+key).Result()
	if err != nil {
		return false, err
	}

	return count == 1, nil
}
