package api

import (
	"fmt"
	"net/http"
)

func getUIDFromRequest(request *http.Request) (int64, error) {
	uid, ok := request.Context().Value(JWTKey("uid")).(int64)
	if !ok {
		return -1, fmt.Errorf("get uid from request context")
	}
	return uid, nil
}
