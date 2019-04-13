package authentication

import (
	"context"
	"net/http"

	"github.com/IWannaCommunity/gate-jump/src/api/res"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	jwt "github.com/dgrijalva/jwt-go"
)

type Level int

const (
	CLAIMS    Level = -1 // just the claims context tag
	PUBLIC    Level = 0  // public return
	USER      Level = 1  // user return
	ADMINUSER Level = 2  // admin wants to change password. admins cant change other users passwords so this exists
	ADMIN     Level = 3  // admin return
	SERVER    Level = 4  // server can update any user without giving 2 shits
)

type Context struct {
	Claims Claims
	Token  string
}

type Claims struct {
	UUID    int64         `json:"uuid"`
	Name    *string       `json:"username"`
	Country *string       `json:"country"`
	Locale  *string       `json:"locale"`
	Group   []interface{} `json:"group"`
	Scope   []interface{} `json:"scope"`
	jwt.StandardClaims
}

func JWTContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := Context{}
		tokenString := r.Header.Get("Authorization")

		if tokenString == "" { // no token provided, value checked will be nil
			ctx := context.WithValue(r.Context(), CLAIMS, nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// parse token provided
		token, err := jwt.ParseWithClaims(tokenString, &claims.Claims,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(settings.JwtSecret), nil
			})
		if err != nil { // token couldn't be read
			res.New(http.StatusUnauthorized).SetErrorMessage("Invalid Token Provided").Error(w)
			return
		}
		if !token.Valid { // token has been edited
			res.New(http.StatusUnauthorized).SetErrorMessage("Token Is Invalid").Error(w)
			return
		}
		if token.Claims == nil { // nothing was put into the token
			res.New(http.StatusInternalServerError).SetErrorMessage("Token Is Null").Error(w)
			return
		}

		ctx := context.WithValue(r.Context(), CLAIMS, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
