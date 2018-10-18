package main

import (
	"context"
	"database/sql"
	"fmt"
	"gate-jump/res"
	"log"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

//shared client,thread-safe: can be used by all request handlers

type Server struct {
	Router  *mux.Router
	DB      *sql.DB
	LogFile *os.File
}

const (
	PUBLIC    = 0
	USER      = 1
	ADMIN     = 2
	ADMINUSER = 3
)

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
	s.Router.Use(s.Recovery)
	//s.Router.Use(s.Logging)
	//s.Router.Use(s.JWTContext)

}

func (s *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(":"+addr, handlers.LoggingHandler(s.LogFile, s.Router)))
}

func (s *Server) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				res.New(http.StatusInternalServerError).SetErrorMessage(fmt.Sprintf("%v", r)).Error(w)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

/*
id==JWTid	admin	context	type
0			0		0		public
0			1		1		user
1			0		2		admin
1			1		3		adminuser
*/

type Claims struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	jwt.StandardClaims
}

func (s *Server) JWTContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("JWTContext Layer Executing")
		claims := Claims{}
		tokenString := r.Header.Get("Authorization")

		if tokenString == "" { // no token provided. public credential only
			ctx := context.WithValue(r.Context(), "AUTH", PUBLIC)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		// parse token provided
		_, err := jwt.ParseWithClaims(tokenString, &claims,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(Config.JwtSecret), nil
			})
		if err != nil { // token couldn't be read. deny completely
			res.New(http.StatusUnauthorized).SetErrorMessage("Invalid Token Provided").Error(w)
			return
		}

		//if claims, ok := token.Claims.(*Claims); ok && token.Valid {

		//}
		next.ServeHTTP(w, r)
	})
}
