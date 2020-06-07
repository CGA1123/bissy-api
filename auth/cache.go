package auth

import (
	"context"
	"time"

	"github.com/cga1123/bissy-api/utils"
	"github.com/go-redis/redis/v8"
)

type CacheStore interface {
	Exists(string) (bool, error)
	Set(string, time.Duration) (string, error)
	Del(string) (bool, error)
	Get(string) (string, error)
}

type RedisStore struct {
	Client      *redis.Client
	IdGenerator utils.IdGenerator
}

func (r *RedisStore) Get(key string) (string, error) {
	value, err := r.Client.Get(context.TODO(), "auth:"+key).Result()
	if err != nil {
		return "", err
	}

	return value, nil
}

func (r *RedisStore) Exists(key string) (bool, error) {
	value, err := r.Client.Exists(context.TODO(), "auth:"+key).Result()
	if err != nil {
		return false, err
	}

	return value == 1, nil
}

func (r *RedisStore) Set(url string, exp time.Duration) (string, error) {
	key := r.IdGenerator.Generate()
	_, err := r.Client.Set(
		context.TODO(),
		"auth:"+key,
		url,
		exp,
	).Result()

	return key, err
}

func (r *RedisStore) Del(key string) (bool, error) {
	count, err := r.Client.Del(context.TODO(), key).Result()
	if err != nil {
		return false, err
	}

	return count == 1, nil
}
