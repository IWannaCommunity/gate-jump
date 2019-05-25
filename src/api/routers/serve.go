package routers

import (
	"fmt"
	"net/http"

	log "github.com/spidernest-go/logger"
	mux "github.com/spidernest-go/mux"
)

var (
	Echo *mux.Echo
)

func Serve(version, port, sslport string) {
	Echo = mux.New()

	prefix := fmt.Sprintf("/oauth/%s/", version)
	owners := prefix + "owners"
	token := prefix + "token"

	Echo.Add(http.MethodGet, prefix, serverInfo)

	Echo.Add(http.MethodPost, owners, createUser)
	Echo.Add(http.MethodPut, owners, updateUser)
	Echo.Add(http.MethodPatch, owners, verifyUser)
	Echo.Add(http.MethodDelete, owners, deleteUser)

	Echo.Add(http.MethodPost, token, createToken)
	Echo.Add(http.MethodPut, token, updateToken)
	Echo.Add(http.MethodDelete, token, deleteToken)

	log.Fatal().Msgf("Router ran into fatal error: %v", Echo.Start(fmt.Sprintf(":%s", port)))
}

func serverInfo(ctx mux.Context) error {
	return mux.ErrNotFound
}
