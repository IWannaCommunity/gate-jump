package routers

import (
    "net/http"

    "github.com/gorilla/mux"
    "github.com/gorilla/handlers"
    "github.com/IWannaCommunity/gate-jump/src/api/log"
    "github.com/IWannaCommunity/gate-jump/src/api/settings"
)

var router *mux.Router

func Serve(port, sslport string) {
    router = mux.NewRouter()

    if settings.Https.CertFile != "" && settings.Https.KeyFile != "" {
		log.Info("HTTPS Credentials picked up, running HTTPS")
		log.Fatal(http.ListenAndServeTLS(":"+sslport,
            settings.Https.CertFile,
            settings.Https.KeyFile,
            handlers.LoggingHandler(s.LogFile, router)
            ))
	} else {
		log.Info("HTTPS Credentials missing, running HTTP")
		log.Fatal(http.ListenAndServe(":"+port, handlers.LoggingHandler(s.LogFile, router)))
	}
}
