package main

import (
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
	config := querycache.Config{
		QueryStore:   querycache.NewInMemoryQueryStore(clock, generator),
		AdapterStore: querycache.NewInMemoryAdapterStore(clock, generator),
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
