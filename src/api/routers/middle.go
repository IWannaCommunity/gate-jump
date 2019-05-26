package routers

import (
	"errors"

	log "github.com/spidernest-go/logger"
	echo "github.com/spidernest-go/mux"
)

// Recover returns a middleware function that returns a handler func that defers a recover handler and passes context into the next route.
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc { // middleware function
		return func(ctx echo.Context) error { // handler function
			defer func() { // recover function
				if r := recover(); r != nil {
					var err error
					switch x := r.(type) {
					case string:
						err = errors.New(x)
					case error:
						err = x
					default: // something really strange happened
						err = errors.New("unknown recover error occurred")
					}
					log.Err(err).Msgf("Recovered successfully from error: %v", err)
					ctx.Error(err)
				}
			}()
			return next(ctx)
		}
	}

}
