package routers

import (
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

	Echo.GET(prefix, serverInfo)

	Echo.POST(owners, createUser)
	Echo.PUT(owners, updateUser)
	Echo.PATCH(owners, verifyUser)
	Echo.DELETE(owners, deleteUser)

	Echo.POST(token, createToken)
	Echo.PUT(token, updateToken)
	Echo.DELETE(token, deleteToken)

	log.Fatal().Msgf("Router ran into fatal error: %v", Echo.Start(fmt.Sprintf(":%s", port)))
}

// serverInfo returns the version information and routes of the server.
func serverInfo(ctx mux.Context) error {
	return ctx.JSON(http.StatusOK, struct {
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
}
