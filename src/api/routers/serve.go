package routers

import (
	"fmt"
	"net/http"

	"github.com/IWannaCommunity/gate-jump/src/api/authentication"
	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/res"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var router *mux.Router

func Serve(port, sslport string) {
	router = mux.NewRouter()

	// Deprecated
	router.HandleFunc("/", getAlive).Methods("GET")
	router.HandleFunc("/user", getUsers).Methods("GET")
	router.HandleFunc("/register", createUser).Methods("POST")
	router.HandleFunc("/login", validateUser).Methods("POST")
	router.HandleFunc("/refresh", refreshUser).Methods("POST")
	router.HandleFunc("/user/{id:[0-9]+}", getUser).Methods("GET")
	router.HandleFunc("/user/{id:[0-9]+}", updateUser).Methods("PUT")
	router.HandleFunc("/user/{id:[0-9]+}", deleteUser).Methods("DELETE")
	//router.HandleFunc("/user/{id:[0-9]+}", banUser).Methods("POST")
	router.HandleFunc("/user/{name}", getUserByName).Methods("GET")
	router.HandleFunc("/verify/{magic}", verifyUser).Methods("GET")
	router.HandleFunc("/scope", createScope).Methods("POST")

	// OAuth 2.0 + OpenID Connect
	router.HandleFunc("/oauth/v1/owner", createOwner).Methods("POST")

	router.Use(HTTPRecovery)
	router.Use(authentication.JWTContext)

	if settings.Https.CertFile != "" && settings.Https.KeyFile != "" {
		log.Info("HTTPS Credentials picked up, running HTTPS")
		log.Fatal(http.ListenAndServeTLS(":"+sslport,
			settings.Https.CertFile,
			settings.Https.KeyFile,
			handlers.LoggingHandler(log.File, router)))
	} else {
		log.Info("HTTPS Credentials missing, running HTTP")
		log.Fatal(http.ListenAndServe(":"+port, handlers.LoggingHandler(log.File, router)))
	}

}

func HTTPRecovery(next http.Handler) http.Handler {
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

func GetRouter() *mux.Router {
	return router
}
