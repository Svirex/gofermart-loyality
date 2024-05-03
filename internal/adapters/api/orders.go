package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Svirex/gofermart-loyality/internal/core/ports"
)

func (api *API) CreateOrder(response http.ResponseWriter, request *http.Request) {
	contentType := request.Header.Get("Content-Type")
	if contentType != "text/plain" {
		api.logger.Debugf("api orders, create order, invalid content type: %v", request)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil || len(body) == 0 {
		api.logger.Error("api orders, create order, read body: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	uid, err := getUIDFromRequest(request)
	if err != nil {
		api.logger.Debugf("api orders, create order, get uid: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := api.ordersService.CreateOrder(request.Context(), uid, string(body))
	if err != nil {
		if errors.Is(err, ports.ErrInvalidOrderNum) {
			api.logger.Debugf("api orders, create order, service response, invalid number: %v", err)
			response.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		api.logger.Debugf("api orders, create order, service response: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	if status == ports.AlreadyAdded {
		response.WriteHeader(http.StatusOK)
		return
	} else if status == ports.Ok {
		response.WriteHeader(http.StatusAccepted)
		return
	} else if status == ports.NotOwnOrder {
		response.WriteHeader(http.StatusConflict)
		return
	} else {
		api.logger.Error("api orders, create order, unknown service response: %v", status)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) GetOrders(response http.ResponseWriter, request *http.Request) {
	uid, err := getUIDFromRequest(request)
	if err != nil {
		api.logger.Debugf("api orders, get orders, get uid: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	orders, err := api.ordersService.GetOrders(request.Context(), uid)
	if err != nil {
		api.logger.Debugf("api orders, get orders, service response: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(orders) == 0 {
		response.WriteHeader(http.StatusNoContent)
		return
	}
	data, err := json.Marshal(orders)
	if err != nil {
		api.logger.Debugf("api orders, get orders, marshal: %v", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	response.Write(data)
}
