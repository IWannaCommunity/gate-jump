package authentication

import (
	"context"
	"net/http"

	"github.com/IWannaCommunity/gate-jump/src/api/res"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	jwt "github.com/dgrijalva/jwt-go"
)

type ContextKey int
type Level int

// Auth Levels
const (
	PUBLIC    Level = 0 // public return
	USER      Level = 1 // user return
	ADMINUSER Level = 2 // admin wants to change password. admins cant change other users passwords so this exists
	ADMIN     Level = 3 // admin return
	SERVER    Level = 4 // server has mega powerlevel
)

// Context Keys
const (
	ClaimsKey ContextKey = -1 // claims context tag
)

type Refresh struct {
	UUID  *string  `json:"uuid"`
	Group []string `json:"group"`
	jwt.StandardClaims
}

type Bearer struct {
	UUID    *string  `json:"uuid"`
	Name    *string  `json:"username"`
	Country *string  `json:"country"`
	Locale  *string  `json:"locale"`
	Group   []string `json:"group"`
	Scope   []string `json:"scope"`
	jwt.StandardClaims
}

func JWTContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var ctx context.Context
		var err error
		var token *jwt.Token
		tokenString := r.Header.Get("Authorization")

		if tokenString == "" { // no token provided, value checked will be nil
			ctx := context.WithValue(r.Context(), ClaimsKey, nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		} else if r.URL.Path == "/refresh" {
			// parse token provided
			claims := Refresh{}
			token, err = jwt.ParseWithClaims(tokenString, &claims,
				func(token *jwt.Token) (interface{}, error) {
					return []byte(settings.JwtSecret), nil
				})
			ctx = context.WithValue(r.Context(), ClaimsKey, claims)
		} else {
			claims := Bearer{}
			token, err = jwt.ParseWithClaims(tokenString, &claims,
				func(token *jwt.Token) (interface{}, error) {
					return []byte(settings.JwtSecret), nil
				})
			ctx = context.WithValue(r.Context(), ClaimsKey, claims)
		}

		if err != nil { // token couldn't be read
			res.New(http.StatusUnauthorized).SetErrorMessage("Invalid Token Provided").Error(w)
			return
		}
		if !token.Valid { // token has been edited
			res.New(http.StatusUnauthorized).SetErrorMessage("Token Is Invalid").Error(w)
			return
		}
		if token.Claims == nil { // nothing was put into the token??
			res.New(http.StatusInternalServerError).SetErrorMessage("Token Is Null").Error(w)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
