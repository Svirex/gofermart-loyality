package api

import (
	"context"
	"errors"
	"net/http"
	"slices"
)

type JWTKey string

func (api *API) CookieAuth(whitelist []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(response http.ResponseWriter, request *http.Request) {
			path := request.URL.Path
			if path[0] != '/' {
				path = "/" + path
			}
			if slices.Contains(whitelist, path) {
				next.ServeHTTP(response, request)
				return
			}
			jwtKey, err := request.Cookie("jwt")
			if errors.Is(err, http.ErrNoCookie) {
				response.WriteHeader(http.StatusUnauthorized)
				return
			}
			uid, err := api.authService.GetUserGromJWT(request.Context(), jwtKey.Value)
			if err != nil {
				api.logger.Errorf("cookie auth middlware, get user from jwt: %v", err)
				response.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(request.Context(), JWTKey("uid"), uid)
			next.ServeHTTP(response, request.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
