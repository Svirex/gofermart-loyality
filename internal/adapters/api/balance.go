package api

import (
	"encoding/json"
	"net/http"
)

func (api *API) GetBalance(response http.ResponseWriter, request *http.Request) {
	uid, err := getUIDFromRequest(request)
	if err != nil {
		api.logger.Debugf("api balance, get balance, get uid: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	balance, err := api.balanceService.GetBalance(request.Context(), uid)
	if err != nil {
		api.logger.Debugf("api balance, get balance, service response: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(balance)
	if err != nil {
		api.logger.Debugf("api balance, get balance, marshal: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.Write(body)
}
