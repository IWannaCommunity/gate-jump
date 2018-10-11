package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mongodb/mongo-go-driver/mongo"
)

//shared client,thread-safe: can be used by all request handlers
var client *mongo.Client

func main() {

	initLog(0, 0, 0) // major,patch,minor
	LoadConfig("/config/config.json")

	log.Printf("Welcome to gate-jump server! Setting up environment...")

	//setup database
	var err error
	client, err = mongo.NewClient("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO()) //TODO: https://golang.org/pkg/context/
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Database initialized!")

	//setup http server
	router := mux.NewRouter()
	router.HandleFunc("/", HomeHandler)
	router.HandleFunc("/register", UserHandler)
	router.HandleFunc("/user/{id}", UserHandler)
	handler := handlers.RecoveryHandler()(router)

	//from https://github.com/gorilla/mux, "Graceful Shutdown"
	srv := &http.Server{
		Addr:         "0.0.0.0:10420",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler,
	}
	go func() {
		log.Printf("Starting Server! Ctrl+C to quit.")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down...")
}

//HomeHandler handles requests to the root
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `{"alive": true}`)
}
