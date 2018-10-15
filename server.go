package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//shared client,thread-safe: can be used by all request handlers

type Server struct {
	Router *mux.Router
	DB     *sql.DB
}

func (s *Server) Initialize(user, password, dbname string) {
	var err error

	s.DB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&interpolateParams=true", user, password, dbname))
	if err != nil {
		log.Fatal(err)
	}

	s.Router = mux.NewRouter()
}

func (s *Server) InitializeRoutes() {
	s.Router.HandleFunc("/", s.getAlive).Methods("GET")
	s.Router.HandleFunc("/user", s.getUsers).Methods("GET")
	s.Router.HandleFunc("/register", s.createUser).Methods("POST")
	s.Router.HandleFunc("/login", s.validateUser).Methods("POST")
	s.Router.HandleFunc("/user/{id}", s.getUser).Methods("GET")
	s.Router.HandleFunc("/user/{id}", s.updateUser).Methods("PUT")
	s.Router.HandleFunc("/user/{id}", s.deleteUser).Methods("DELETE")
}

func (s *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(":"+addr, s.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	log.Println(fmt.Sprintf("Error Code: %d; Error Message: %s", code, message))
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
