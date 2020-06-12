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
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
	"github.com/jmoiron/sqlx"
)

const (
	REDIS_URL            = "REDISCLOUD_URL"
	JWT_SIGNING_KEY      = "JWT_SIGNING_KEY"
	GITHUB_CLIENT_ID     = "GITHUB_CLIENT_ID"
	GITHUB_CLIENT_SECRET = "GITHUB_CLIENT_SECRET"
	DATABASE_URL         = "DATABASE_URL"
	PORT                 = "PORT"
)

func initAuth(env map[string]string, db *hnysqlx.DB, redis *redis.Client) *auth.Config {
	clock := &utils.RealClock{}
	idGen := &utils.UUIDGenerator{}

	store := auth.NewSQLUserStore(db, clock, idGen)

	return auth.NewConfig(
		[]byte(env[JWT_SIGNING_KEY]),
		store,
		clock,
		&auth.RedisStore{Client: redis, IdGenerator: idGen},
		auth.NewGithubApp(env[GITHUB_CLIENT_ID], env[GITHUB_CLIENT_SECRET], &http.Client{}),
	)
}

func initRedis(vars map[string]string) *redis.Client {
	options, err := redis.ParseURL(vars[REDIS_URL])
	if err != nil {
		log.Fatalf("failed to parse redis url %v", err)
	}

	// rediscloud sets user:password but rediscloud is not ACL AUTH enabled
	options.Username = ""

	redisClient := redis.NewClient(options)
	if err := redisClient.Ping(context.TODO()).Err(); err != nil {
		log.Fatalf("failed to ping redis %v", err)
	}

	return redisClient
}

func initHoneycomb() {
	beeline.Init(beeline.Config{
		WriteKey:    os.Getenv("HONEYCOMB_WRITEKEY"),
		Dataset:     "bissy-api",
		ServiceName: "bissy-api",
	})

}

func initDb(env map[string]string) *hnysqlx.DB {
	db, err := sqlx.Open("postgres", env[DATABASE_URL])
	if err != nil {
		log.Fatalf("failed to open db %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db %v", err)
	}

	return hnysqlx.WrapDB(db)
}

func homeHandler(w http.ResponseWriter, h *http.Request) {
	fmt.Fprintf(w, "bissy-api")
}

func main() {
	env, err := utils.RequireEnv(
		PORT,
		DATABASE_URL,
		JWT_SIGNING_KEY,
		GITHUB_CLIENT_ID,
		GITHUB_CLIENT_SECRET,
		REDIS_URL,
	)
	if err != nil {
		log.Fatal(err)
	}

	clock := &utils.RealClock{}
	generator := &utils.UUIDGenerator{}
	initHoneycomb()
	redisClient := initRedis(env)
	db := initDb(env)
	authConfig := initAuth(env, db, redisClient)

	router := mux.NewRouter()
	router.Use(hnygorilla.Middleware)
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/ping", ping.Handler)
	router.Handle("/authping", authConfig.Middleware(http.HandlerFunc(ping.Handler)))

	authMux := router.PathPrefix("/auth").Subrouter()
	authConfig.SetupHandlers(authMux)

	querycacheMux := router.PathPrefix("/querycache").Subrouter()
	queryCacheConfig := querycache.Config{
		QueryStore:      querycache.NewSQLQueryStore(db, clock, generator),
		DatasourceStore: querycache.NewSQLDatasourceStore(db, clock, generator),
		Cache:           &querycache.RedisCache{Client: redisClient},
		Clock:           clock,
	}

	querycacheMux.Use(authConfig.Middleware)
	queryCacheConfig.SetupHandlers(querycacheMux)

	handler := handlers.LoggingHandler(os.Stdout, hnynethttp.WrapHandler(router))

	server := &http.Server{
		Addr:         "0.0.0.0:" + env[PORT],
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
