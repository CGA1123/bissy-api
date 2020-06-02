package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cga1123/bissy-api/ping"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/wrappers/hnygorilla"
	"github.com/honeycombio/beeline-go/wrappers/hnynethttp"
)

func initRedis() (*redis.Client, error) {
	redisUrl, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		redisUrl = "redis://localhost:6379"
	}

	options, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(options)
	if err := redisClient.Ping(context.TODO()).Err(); err != nil {
		return nil, err
	}

	return redisClient, nil
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	hcWrite := os.Getenv("HONEYCOMB_WRITEKEY")
	beeline.Init(beeline.Config{
		WriteKey:    hcWrite,
		Dataset:     "bissy-api",
		ServiceName: "bissy-api",
	})

	redisClient, err := initRedis()
	if err != nil {
		log.Fatal(err)
	}

	cache := &querycache.RedisCache{Client: redisClient, Prefix: "querycache"}

	router := mux.NewRouter()
	router.Use(hnygorilla.Middleware)
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
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      hnynethttp.WrapHandler(router),
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
