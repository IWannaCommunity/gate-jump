package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mongodb/mongo-go-driver/mongo"
)

//shared client,thread-safe: can be used by all request handlers
var client *mongo.Client

func main() {
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

//UserHandler handles requests to /user/{id}
func UserHandler(w http.ResponseWriter, r *http.Request) {
	//get userid from path as integer
	vars := mux.Vars(r)
	useridStr := vars["id"]
	userid, err := strconv.Atoi(useridStr)
	if err != nil { //if userid can't be converted to integer, tell the client it's WRONG
		http.Error(w, "Invalid user id", 400)
		return
	}

	//get user from database
	collection := client.Database("gatejump").Collection("users")
	result := UserObject{}
	err = collection.FindOne(context.Background(), map[string]int{"userid": userid}).Decode(&result)
	if err == mongo.ErrNoDocuments { //if no user found, return 404
		http.Error(w, "No user", 404)
		return
	} else if err != nil { //otherwise, panic with the error (500)
		panic(err)
	}

	jsonResult, err := json.Marshal(result)
	if err != nil { //if can't marshal as json, panic with the error (500)
		panic("Unable to encode json")
	}

	//all good, tell the client json is incoming, and send it the json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonResult)
}

//UserObject is a representation of the user object in the database
type UserObject struct {
	Userid      int
	Username    string
	Password    string `json:"-"` //don't share the password via json responses
	Email       string
	Country     string
	DateCreated time.Time `bson:"dateCreated"`
	Verified    bool
	Banned      bool
	LastToken   string    `bson:"lastToken"`
	LastLogin   time.Time `bson:"lastLogin"`
	LastIP      string    `bson:"lastIP" json:"-"` //don't share the IP via json responses
}
