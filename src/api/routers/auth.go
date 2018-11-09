package routers

import (
    "net/http"
    "context"

    jwt "github.com/dgrijalva/jwt-go"
    "github.com/IWannaCommunity/gate-jump/src/api/res"
    "github.com/IWannaCommunity/gate-jump/src/api/settings"
)

type AuthLevel int

const (
	CLAIMS    AuthLevel = -1 // just the claims context tag
	PUBLIC    AuthLevel = 0  // public return
	USER      AuthLevel = 1  // user return
	ADMINUSER AuthLevel = 2  // admin wants to change password. admins cant change other users passwords so this exists
	ADMIN     AuthLevel = 3  // admin return
	SERVER    AuthLevel = 4  // server can update any user without giving 2 shits
)

type Context struct {
	claims Claims
	token  string
}

type Claims struct {
	ID       int64   `json:"id"`
	Name     *string `json:"username"`
	Admin    bool    `json:"admin"`
	Country  *string `json:"country"`
	Locale   *string `json:"locale"`
	Verified bool    `json:"verified"`
	Banned   bool    `json:"banned"`
	jwt.StandardClaims
}

func JWTContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextData := Context{}
		tokenString := r.Header.Get("Authorization")

		if tokenString == "" { // no token provided. public credential only
			ctx := context.WithValue(r.Context(), CLAIMS, Context{claims: Claims{ID: 0}})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		// parse token provided
		token, err := jwt.ParseWithClaims(tokenString, &contextData.claims,
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
		contextData.token = tokenString
		contextData.claims = *token.Claims.(*Claims)

		ctx := context.WithValue(r.Context(), CLAIMS, contextData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
