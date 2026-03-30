package router

import (
	"avgys-gophermat/internal/endpoints"
	"avgys-gophermat/internal/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const textType = "text/plain"
const xgzipType = "application/x-gzip"
const jsonType = "application/json"

func NewRouter(h *endpoints.Endpoints) *chi.Mux {

	r := chi.NewRouter()

	r.Use(middleware.RealIP, middlewares.WithLogging, middlewares.WithCompression)

	r.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {

			r.Use(middleware.AllowContentType(jsonType))

			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
		})

		r.Route("/orders", func(r chi.Router) {
			r.Use(middlewares.RequireCookie)

			r.Post("/", h.LoadOrder)
			r.Get("/", h.GetOrdersByUserId)
		})
	})

	return r
}
