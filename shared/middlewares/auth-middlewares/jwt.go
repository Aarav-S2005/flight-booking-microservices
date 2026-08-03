package auth_middlewares

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

func Verifier(tokenAuth *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return jwtauth.Verifier(tokenAuth)
}

func Authenticator(tokenAuth *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return jwtauth.Authenticator(tokenAuth)
}
