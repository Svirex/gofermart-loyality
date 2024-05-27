package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Svirex/gofermart-loyality/internal/core/ports"
)

type authData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (api *API) Register(response http.ResponseWriter, request *http.Request) {
	contentType := request.Header.Get("Content-Type")
	if contentType != "application/json" {
		api.logger.Debugf("api auth, register, invalid content type: %v", request)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil || len(body) == 0 {
		api.logger.Error("api auth, register, read body: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	request.Body.Close()
	var auth authData
	err = json.Unmarshal(body, &auth)
	if err != nil {
		api.logger.Error("api auth, register, unmarshal: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	jwt, err := api.authService.Register(request.Context(), auth.Login, auth.Password)
	if err != nil {
		if errors.Is(err, ports.ErrUserAlreadyExists) {
			response.WriteHeader(http.StatusConflict)
			return
		}
		api.logger.Error("api auth, register, service response: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	jwtCookie := &http.Cookie{
		Name:  "jwt",
		Value: jwt,
	}
	http.SetCookie(response, jwtCookie)
}

func (api *API) Login(response http.ResponseWriter, request *http.Request) {
	contentType := request.Header.Get("Content-Type")
	if contentType != "application/json" {
		api.logger.Debugf("api auth, login, invalid content type: %v", request)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil || len(body) == 0 {
		api.logger.Error("api auth, login, read body: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	request.Body.Close()
	var auth authData
	err = json.Unmarshal(body, &auth)
	if err != nil {
		api.logger.Error("api auth, login, unmarshal: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	jwt, err := api.authService.Login(request.Context(), auth.Login, auth.Password)
	if err != nil {
		if errors.Is(err, ports.ErrUserNotFound) || errors.Is(err, ports.ErrInvalidPassword) {
			api.logger.Debugf("api auth, login, read body, service error: %v", err)
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		api.logger.Error("api auth, login, service response: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	jwtCookie := &http.Cookie{
		Name:  "jwt",
		Value: jwt,
	}
	http.SetCookie(response, jwtCookie)
}
