package handler

import (
	shared "avgys-gophermat/internal/shared/http"
	"net/http"
)

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	// traceLogger := logger.Endpoint(r.Context(), "ping db")

	err := h.Store.TestConnection(r.Context())

	if err != nil {

		shared.WriteResponse(w, nil, http.StatusInternalServerError)
	}

	shared.WriteResponse(w, nil, http.StatusOK)
}
