package routers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/IWannaCommunity/gate-jump/src/api/settings"
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

// STB Converts a given structure to a byte array.
// Setup as a function so it can be improved upon later.
func STB(s interface{}) ([]byte, error) {
	marshaled, err := json.Marshal(s)
	if err != nil {
		return []byte(""), err
	}
	return []byte(marshaled), nil
}

func serverInfo(ctx mux.Context) error {
	r, err := STB(
		struct {
			Minor  int          `json:"minor"`
			Patch  int          `json:"patch"`
			Major  int          `json:"major"`
			Routes []*mux.Route `json:"routes"`
		}{
			settings.Minor,
			settings.Patch,
			settings.Major,
			Echo.Routes(),
		})
	if err != nil {
		log.Err(err).Msgf("Error converting struct to byte array: %v", err)
		return mux.ErrInternalServerError
	}
	_, err = mux.NewResponse(ctx.Request(), ctx.Echo()).Write(r)
	if err != nil {
		log.Err(err).Msgf("Unknown error sending server info: %v", err)
		return mux.ErrInternalServerError
	}
	return nil
}
