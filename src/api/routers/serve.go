package routers

import (
    "fmt"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/gorilla/handlers"
    "github.com/IWannaCommunity/gate-jump/src/api/log"
    "github.com/IWannaCommunity/gate-jump/src/api/settings"
    "github.com/IWannaCommunity/gate-jump/src/api/res"
    "github.com/IWannaCommunity/gate-jump/src/api/authentication"
)

var router *mux.Router

func Serve(port, sslport string) {
    router = mux.NewRouter()

    router.HandleFunc("/", getAlive).Methods("GET")
    router.HandleFunc("/user", getUsers).Methods("GET")
    router.HandleFunc("/register", createUser).Methods("POST")
    router.HandleFunc("/login", validateUser).Methods("POST")
    router.HandleFunc("/refresh", refreshUser).Methods("POST")
    router.HandleFunc("/user/{id:[0-9]+}", getUser).Methods("GET")
    router.HandleFunc("/user/{id}", getUserByName).Methods("GET")
    router.HandleFunc("/user/{id:[0-9]+}", updateUser).Methods("PUT")
    router.HandleFunc("/user/{id:[0-9]+}", deleteUser).Methods("DELETE")
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
