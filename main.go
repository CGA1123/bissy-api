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
	"github.com/cga1123/bissy-api/auth/github"
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
	"github.com/rs/cors"
)

const (
	redisURLVar           = "REDISCLOUD_URL"
	jwtSigningKeyVar      = "JWT_SIGNING_KEY"
	githubClientIDVar     = "GITHUB_CLIENT_ID"
	githubClientSecretVar = "GITHUB_CLIENT_SECRET"
	databaseURLVar        = "DATABASE_URL"
	portVar               = "PORT"
	frontendOrigin        = "FRONTEND_ORIGIN"
)

func initCors(frontend string) *cors.Cors {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodHead,
	}

	return cors.New(cors.Options{
		AllowCredentials: true,
		AllowedMethods:   methods,
		AllowedHeaders:   []string{"Authorization"},
		AllowedOrigins:   []string{frontend}})
}

func initRedis(vars map[string]string) *redis.Client {
	options, err := redis.ParseURL(vars[redisURLVar])
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
	db, err := sqlx.Open("postgres", env[databaseURLVar])
	if err != nil {
		log.Fatalf("failed to open db %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db %v", err)
	}

	return hnysqlx.WrapDB(db)
}

func initQueryCache(db *hnysqlx.DB, clock utils.Clock, gen utils.IDGenerator, redisClient *redis.Client) *querycache.Config {
	return &querycache.Config{
		QueryStore:      querycache.NewSQLQueryStore(db, clock, gen),
		DatasourceStore: querycache.NewSQLDatasourceStore(db, clock, gen),
		Cache:           &querycache.RedisCache{Client: redisClient},
		Clock:           clock,
	}
}

func requireEnv() map[string]string {
	env, err := utils.RequireEnv(
		redisURLVar,
		jwtSigningKeyVar,
		githubClientIDVar,
		githubClientSecretVar,
		databaseURLVar,
		portVar,
		frontendOrigin,
	)

	if err != nil {
		log.Fatal(err)
	}

	return env
}

func homeHandler(w http.ResponseWriter, h *http.Request) {
	fmt.Fprintf(w, "bissy-api")
}

func runServer(handler http.Handler, port string) *http.Server {
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

	return server
}

func shutdown(server *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	log.Println("shutting down")
}

func main() {
	env := requireEnv()
	clock := &utils.RealClock{}
	generator := &utils.UUIDGenerator{}
	initHoneycomb()
	redisClient := initRedis(env)
	db := initDb(env)
	authConfig := auth.New([]byte(env[jwtSigningKeyVar]))
	corsConfig := initCors(env[frontendOrigin])

	router := mux.NewRouter()
	router.Use(hnygorilla.Middleware, corsConfig.Handler)
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/ping", ping.Handler)
	router.Handle("/authping", authConfig.Middleware(http.HandlerFunc(ping.Handler)))

	githubAuthMux := router.PathPrefix("/auth/github").Subrouter()
	githubApp := github.NewApp(env[githubClientIDVar], env[githubClientSecretVar], &http.Client{Timeout: time.Second * 5})
	githubAuthConfig := github.New(authConfig, db, redisClient, githubApp)
	githubAuthConfig.SetupHandlers(githubAuthMux)

	queryCacheConfig := initQueryCache(db, clock, generator, redisClient)
	querycacheMux := router.PathPrefix("/querycache").Subrouter()
	querycacheMux.Use(authConfig.Middleware)
	queryCacheConfig.SetupHandlers(querycacheMux)

	handler := handlers.LoggingHandler(os.Stdout, hnynethttp.WrapHandler(router))

	shutdown(runServer(handler, env[portVar]))

	os.Exit(0)
}
