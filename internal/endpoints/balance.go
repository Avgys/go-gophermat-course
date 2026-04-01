package endpoints

import (
	"avgys-gophermat/internal/logger"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	"encoding/json"
	"net/http"
)

func (e *Endpoints) GetBalanceByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "GetBalanceByUserId")

	userClaims, err := auth.GetFromContext(ctx)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	orders, err := e.BalanceService.GetBalanceByUserID(ctx, userClaims)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	body, err := json.Marshal(orders)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	httphelper.WriteResponse(w, body, http.StatusOK)
}
