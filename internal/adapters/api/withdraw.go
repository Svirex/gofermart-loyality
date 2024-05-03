package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
)

func (api *API) Withdraw(response http.ResponseWriter, request *http.Request) {
	contentType := request.Header.Get("Content-Type")
	if contentType != "application/json" {
		api.logger.Debugf("api withdraw, withdraw, invalid content type: %v", request)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil || len(body) == 0 {
		api.logger.Error("api withdraw, withdraw, read body: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	request.Body.Close()
	data := &domain.WithdrawData{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		api.logger.Error("api withdraw, withdraw, unmarshal: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	uid, err := getUIDFromRequest(request)
	if err != nil {
		api.logger.Debugf("api withdraw, withdraw, get uid: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = api.withdrawService.Withdraw(request.Context(), uid, data)
	if err != nil {
		if errors.Is(err, ports.ErrInvalidOrderNum) || errors.Is(err, ports.ErrDuplicateOrderNumber) {
			response.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, ports.ErrNotEnoughMoney) {
			response.WriteHeader(http.StatusPaymentRequired)
			return
		}
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) GetWithdrawals(response http.ResponseWriter, request *http.Request) {
	uid, err := getUIDFromRequest(request)
	if err != nil {
		api.logger.Debugf("api withdraw, get withdrawals, get uid: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := api.withdrawService.GetWithdrawals(request.Context(), uid)
	if err != nil {
		api.logger.Debugf("api withdraw, get withdrawals, service response: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(data) == 0 {
		response.WriteHeader(http.StatusNoContent)
		return
	}
	body, err := json.Marshal(data)
	if err != nil {
		api.logger.Debugf("api withdraw, get withdrawals, service response: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.Write(body)
}
