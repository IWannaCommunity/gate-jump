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

func Serve(version, port, sslport string) {
	router = mux.NewRouter()

	prefix := fmt.Sprintf("/oauth/%s/", version)
	owners := prefix + "owners"
	token := prefix + "token"

	router.HandleFunc(prefix, serverInfo).Methods("GET")
	router.HandleFunc(owners, createUser).Methods("POST")
	router.HandleFunc(owners, updateUser).Methods("PUT")
	router.HandleFunc(owners, verifyUser).Methods("PATCH")
	router.HandleFunc(owners, deleteUser).Methods("DELETE")
	router.HandleFunc(token, createToken).Methods("POST")
	router.HandleFunc(token, updateToken).Methods("PUT")
	router.HandleFunc(token, deleteToken).Methods("DELETE")

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
