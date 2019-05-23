package routers

import (
	"fmt"
	"net/http"

	log "github.com/spidernest-go/logger"
	mux "github.com/spidernest-go/mux"
)

var (
	Echo   *mux.Echo
	Router *mux.Router
)

func Serve(version, port, sslport string) {
	Echo = mux.New()
	Router = Echo.Router()

	prefix := fmt.Sprintf("/oauth/%s/", version)
	owners := prefix + "owners"
	token := prefix + "token"

	Router.Add(http.MethodGet, prefix, serverInfo)

	Router.Add(http.MethodPost, owners, createUser)
	Router.Add(http.MethodPut, owners, updateUser)
	Router.Add(http.MethodPatch, owners, verifyUser)
	Router.Add(http.MethodDelete, owners, deleteUser)

	Router.Add(http.MethodPost, token, createToken)
	Router.Add(http.MethodPut, token, updateToken)
	Router.Add(http.MethodDelete, token, deleteToken)

	log.Fatal().Msgf("Router ran into fatal error: %v", Echo.Start(fmt.Sprintf(":%s", port)))
}

func serverInfo(ctx mux.Context) error {
	return mux.ErrNotFound
}
