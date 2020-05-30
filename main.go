package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cga1123/bissy-api/ping"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/gorilla/mux"
)

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	router := mux.NewRouter()
	router.HandleFunc("/ping", ping.Handler)

	querycacheMux := router.PathPrefix("/querycache").Subrouter()
	clock := &querycache.RealClock{}
	generator := &querycache.UUIDGenerator{}
	executor, err := querycache.NewSQLExecutor("postgres", "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	config := querycache.Config{
		Store:    querycache.NewInMemoryQueryStore(clock, generator),
		Executor: executor,
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
