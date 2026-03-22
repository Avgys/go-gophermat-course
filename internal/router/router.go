package router

import (
	"avgys-gophermat/internal/handler"
	"avgys-gophermat/internal/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const textType = "text/plain"
const xgzipType = "application/x-gzip"
const jsonType = "application/json"

func NewRouter(h *handler.Handlers) *chi.Mux {

	r := chi.NewRouter()
	setEndpoints(r, h)

	return r
}

func setEndpoints(r *chi.Mux, h *handler.Handlers) {

	r.Use(middleware.RealIP, middlewares.WithLogging, middlewares.WithCompression)

	r.Get("/ping", h.Ping)
}
