package endpoints

import (
	"avgys-gophermat/internal/logger"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	"encoding/json"
	"net/http"
)

func (e *Endpoints) LoadOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "LoadOrder")

	userClaims, err := auth.GetFromContext(ctx)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	body, err := getBody(w, r)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	orderNum := string(body)

	err = e.OrderService.Store(ctx, userClaims, orderNum)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	httphelper.WriteResponse(w, nil, http.StatusOK)
}

func (e *Endpoints) GetOrdersByUserId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "GetOrdersByUserId")

	userClaims, err := auth.GetFromContext(ctx)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	orders, err := e.OrderService.GetOrderByUserID(ctx, userClaims)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	body, err := json.Marshal(orders)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	w.Header().Set("Content-type", "application/json")
	httphelper.WriteResponse(w, body, http.StatusOK)
}
