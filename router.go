package main

import (
	"context"
	"io"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

// AuthLevel int representation of allowable authentication level
type AuthLevel int

// Routes array of routes for iterating through
type Routes []Route

// HandlerFunc handler function format
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// AuthLevel Int Value declaration
const (
	PUBLIC = 0
	USER   = 1
	ADMIN  = 2
)

// Route the route data structure
type Route struct {
	Method   string
	Pattern  string
	Function HandlerFunc
	Auth     AuthLevel
	Name     string
}

// Claims the claim data structure
type Claims struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	jwt.StandardClaims
}

// UserRequest helper structure
type UserRequest struct {
	http.Request
}

// Home simplistic testing function for seeing if server is up
func Home(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `{"alive": true}`)
}

// Defined Routes for the API and their relevant information
var routes = Routes{
	Route{"GET", "/", Home, PUBLIC, "home"},
	Route{"GET", "/user/{id}", GetUser, PUBLIC, "user"},
}

// NewRouter returns a router with all given routes
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	//Schemes("https") //BUG(klazen108): enable this when ready to go https only
	//Host("sfm.delicious-fruit.com") //do we want this?

	for _, route := range routes {
		log.Printf(
			"Adding route: %6s %s -> %s",
			route.Method, route.Pattern, GetFunctionName(route.Function),
		)

		var h http.Handler
		h = http.HandlerFunc(route.Function)
		h = CheckAuthLevel(h, route.Auth)
		h = Logger(h, route.Name)
		//TODO: if RouteBase is just a /, then don't prepend it
		router.
			Methods(route.Method).
			Path(Config.RouteBase + route.Pattern).
			Name(route.Name).
			Handler(h)
	}

	return router
}

// CheckAuthLevel ?
func CheckAuthLevel(inner http.Handler, auth AuthLevel) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Check claims, if they exist
		authenticated := false
		claims := Claims{}

		tokenString := r.Header.Get("Authorization")
		if tokenString != "" {
			token, err := jwt.ParseWithClaims(tokenString, &claims,
				func(token *jwt.Token) (interface{}, error) {
					return []byte(Config.JwtSecret), nil
				})
			if err != nil {
				httpResponse(w, "Unauthorized", 401)
				return
			}

			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				authenticated = true
				//BUG(klazen108): refactor to pass claims in a type-safe way
				ctx := context.WithValue(r.Context(), 0, claims)
				//at this point claims has all the info about the user, need to pass
				//this in to the underlying handlers
				r = r.WithContext(ctx)
			}
		}
		//r = UserRequest{r,claims}

		switch auth {
		case PUBLIC: //serve the request regardless
			inner.ServeHTTP(w, r)
		case USER: //fail if there were no valid claims
			if !authenticated {
				httpResponse(w, "Unauthorized", 401)
				return
			}
			inner.ServeHTTP(w, r)
		case ADMIN: //fail if invalid claims or valid claims, but user is not admin
			if !authenticated || !claims.Admin {
				httpResponse(w, "Unauthorized", 401)
				return
			}
			inner.ServeHTTP(w, r)
		}
	})
}

// GetClaims gest the claims from the request
func (req UserRequest) GetClaims() *Claims {
	claim, ok := req.Context().Value(0).(*Claims) // r u ok?
	if !ok {                                      // im not okay man
		return nil
	}
	return claim
}
