package endpoints

import (
	"avgys-gophermat/internal/logger"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	"net/http"
)

func (e *Endpoints) LoadOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "LoadOrder")
	userClaims, err := auth.GetFromContext(ctx)

	if err != nil {
		httphelper.WriteError(w, r, err, traceLogger)
		return
	}

	body, err := getBody(w, r)

	if err != nil {
		httphelper.WriteError(w, r, err, traceLogger)
		return
	}

	orderNum := string(body)

	err = e.OrderService.Store(ctx, userClaims, orderNum)

	if err != nil {
		httphelper.WriteError(w, r, err, traceLogger)
		return
	}

	httphelper.WriteResponse(w, nil, http.StatusOK)
}
