package main

import (
	"net/http"
	"os"
	"time"

	"github.com/cga1123/bissy-api/ping"
	"github.com/cga1123/bissy-api/robert"
	"github.com/gorilla/mux"
)

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	router := mux.NewRouter()
	router.HandleFunc("/ping", ping.Handler)

	robertMux := router.PathPrefix("/robert").Subrouter()
	clock := &robert.RealClock{}
	generator := &robert.UUIDGenerator{}
	config := robert.Config{
		Store: robert.NewInMemoryStore(clock, generator),
	}

	config.SetupHandlers(robertMux)

	server := &http.Server{
		Addr:         ":" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	server.ListenAndServe()
}
