package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	log.Printf("Welcome to gate-jump server! Ctrl+C to quit.")

	router := mux.NewRouter()
	router.HandleFunc("/", HomeHandler)
	handler := handlers.RecoveryHandler()(router)

	log.Fatal(http.ListenAndServe(":10420", handler))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `{"alive": true}`)
}
