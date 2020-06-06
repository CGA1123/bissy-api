package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/ping"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/cga1123/bissy-api/utils"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/wrappers/hnygorilla"
	"github.com/honeycombio/beeline-go/wrappers/hnynethttp"
	"github.com/jmoiron/sqlx"
)

func initAuth(db *sqlx.DB, redis *redis.Client) (*auth.Config, error) {
	signingKey, ok := os.LookupEnv("JWT_SIGNING_KEY")
	if !ok {
		return nil, fmt.Errorf("JWT_SIGNING_KEY not set")
	}

	githubClientId, ok := os.LookupEnv("GITHUB_CLIENT_ID")
	if !ok {
		return nil, fmt.Errorf("GITHUB_CLIENT_ID not set")
	}

	githubClientSecret, ok := os.LookupEnv("GITHUB_CLIENT_SECRET")
	if !ok {
		return nil, fmt.Errorf("GITHUB_CLIENT_ID not set")
	}

	clock := &utils.RealClock{}
	idGen := &utils.UUIDGenerator{}

	store := auth.NewSQLUserStore(db, clock, idGen)

	return auth.NewConfig(
		[]byte(signingKey),
		store,
		clock,
		&auth.RedisStore{Client: redis, IdGenerator: idGen},
		auth.NewGithubApp(githubClientId, githubClientSecret, &http.Client{}),
	), nil
}

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

func initDb() (*sqlx.DB, error) {
	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		databaseUrl = "postgres://localhost:5432"
	}

	db, err := sqlx.Open("postgres", databaseUrl)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
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

	db, err := initDb()
	if err != nil {
		log.Fatal(err)
	}

	clock := &utils.RealClock{}
	generator := &utils.UUIDGenerator{}
	authConfig, err := initAuth(db, redisClient)
	if err != nil {
		log.Fatal("could not init auth:", err)
	}

	router := mux.NewRouter()
	router.Use(hnygorilla.Middleware)
	router.HandleFunc("/ping", ping.Handler)
	router.Handle("/authping", authConfig.WithAuth(http.HandlerFunc(ping.Handler)))

	authMux := router.PathPrefix("/auth").Subrouter()
	authConfig.SetupHandlers(authMux)

	querycacheMux := router.PathPrefix("/querycache").Subrouter()
	queryCacheConfig := querycache.Config{
		QueryStore:   querycache.NewSQLQueryStore(db, clock, generator),
		AdapterStore: querycache.NewSQLAdapterStore(db, clock, generator),
		Cache:        &querycache.RedisCache{Client: redisClient},
		Clock:        clock,
	}

	querycacheMux.Use(authConfig.WithAuth)
	queryCacheConfig.SetupHandlers(querycacheMux)

	handler := handlers.LoggingHandler(os.Stdout, hnynethttp.WrapHandler(router))

	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler,
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
