package main

import (
	"database/sql"

	"github.com/gorilla/mux"
)

//shared client,thread-safe: can be used by all request handlers

type Server struct {
	Router *mux.Router
	DB     *sql.DB
}

func (s *Server) Initialize(user, password, dbname string) {

}

func (s *Server) Run(addr string) {

}
