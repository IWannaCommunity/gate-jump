package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

type Response struct {
	Function string        // name of the function called
	Code     int           // response code returned (determines if logged)
	Message  string        // response message
	Payload  interface{}   // json payload if relevant
	Args     []interface{} // arguments to sql query if relevant
	Query    string        // sql query if relevant
	Err      error         // error if relevant
}

func NewResponse(code int, message string, payload interface{}) Response {
	// predefine any relevant information for normal non-important response
	r := Response{
		Function: MyCaller(),
		Code:     code,
		Message:  message,
		Payload:  payload,
		Args:     nil,
		Query:    "",
		Err:      nil,
	}
	return r
}

func respondWithError(w http.ResponseWriter, r Response) {
	if r.Code >= 500 {
		/*
			Example:
			2018/10/17 21:10:47 ERROR: Failed to get user data from database.
			(500) getUser:	Invalid sql syntax check something something
			({<nil>}): "SELECT * FROM users WHERE id=?"
		*/
		log.Printf("ERROR: %s\n(%d) %s: %v\n(%v): \"%s\"", r.Message, r.Code, r.Function, r.Err, r.Args, r.Query)
	}
	if r.Message == "" {
		r.Message = r.Err.Error()
	}
	r.Payload = map[string]string{"error": r.Message}
	respondWithJSON(w, r)
}

func respondWithJSON(w http.ResponseWriter, r Response) {
	payload, _ := json.Marshal(r.Payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	w.Write(payload)
}

func (s *Server) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from Panic: ", r)
				respondWithError(w, NewResponse(http.StatusInternalServerError, fmt.Sprintf("%v", r), nil))
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
			respondWithError(w, NewResponse(http.StatusUnauthorized, "Invalid token provided", nil))
			return
		}

		//if claims, ok := token.Claims.(*Claims); ok && token.Valid {

		//}
		next.ServeHTTP(w, r)
	})
}
