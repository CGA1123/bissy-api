package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cga1123/bissy-api/ping"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	redisUrl, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		redisUrl = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})

	err := redisClient.Ping(context.TODO()).Err()
	if err != nil {
		log.Fatal(err)
	}

	cache := &querycache.RedisCache{Client: redisClient, Prefix: "querycache"}

	router := mux.NewRouter()
	router.HandleFunc("/ping", ping.Handler)

	querycacheMux := router.PathPrefix("/querycache").Subrouter()
	clock := &querycache.RealClock{}
	generator := &querycache.UUIDGenerator{}
	config := querycache.Config{
		QueryStore:   querycache.NewInMemoryQueryStore(clock, generator),
		AdapterStore: querycache.NewInMemoryAdapterStore(clock, generator),
		Cache:        cache,
		Clock:        clock,
	}

	config.SetupHandlers(querycacheMux)

	server := &http.Server{
		Addr:         ":" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	server.ListenAndServe()
}
