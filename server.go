package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

//shared client,thread-safe: can be used by all request handlers

type Server struct {
	Router *mux.Router
	DB     *sql.DB
}

type Claims struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	jwt.StandardClaims
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

func (s *Server) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Recovery Layer Executing")
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from Panic: ", r)
				respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%v", r))
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Logging Layer Executing")
		ts := time.Now()
		next.ServeHTTP(w, r)
		// ex: GET /users -> 500 (30.00s)
		log.Printf("%s\t%s\t%s\t" /*%d*/, r.Method, r.RequestURI, time.Since(ts) /*,lw.statusCode*/)
	})
}

/*
id==JWTid	admin	context	type
0			0		0		public
0			1		1		user
1			0		2		admin
1			1		3		adminuser
*/

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
			respondWithError(w, http.StatusUnauthorized, "Invalid token provided")
			return
		}

		//if claims, ok := token.Claims.(*Claims); ok && token.Valid {

		//}
		next.ServeHTTP(w, r)
	})
}
