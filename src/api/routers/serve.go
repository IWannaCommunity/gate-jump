package routers

import (
	"fmt"
	"net/http"

	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	mux "github.com/labstack/echo"
	middleware "github.com/labstack/echo/middleware"
	log "github.com/spidernest-go/logger"
	//mux "github.com/spidernest-go/mux"
)

// Serve initalizes the router to the given port, adding the routes to them, middleware, and then starting the server until a fatal error occurs.
func Serve(version, port, sslport string) {
	e := mux.New()

	prefix := fmt.Sprintf("/oauth/%s/", version)
	owners := prefix + "owners"
	token := prefix + "token"

	e.GET(prefix, serverInfo)

	e.POST(owners, createUser)
	e.PUT(owners, updateUser)
	e.PATCH(owners, verifyUser)
	e.DELETE(owners, deleteUser)

	e.POST(token, createToken)
	e.PUT(token, updateToken)
	e.DELETE(token, deleteToken)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	log.Fatal().Msgf("Router ran into fatal error: %v", e.Start(fmt.Sprintf(":%s", port)))
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
		ctx.Echo().Routes(),
	})
}
