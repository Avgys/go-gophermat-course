package router

import (
	"avgys-gophermat/internal/endpoints"
	"avgys-gophermat/internal/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const jsonType = "application/json"

func NewRouter(h *endpoints.Endpoints) *chi.Mux {

	r := chi.NewRouter()

	r.Use(middlewares.WithLogging, middleware.Recoverer, middleware.RealIP, middlewares.WithCompression)

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

		r.Route("/balance", func(r chi.Router) {
			r.Use(middlewares.RequireCookie)

			r.Get("/", h.GetBalanceByUserID)
			r.With(middleware.AllowContentType(jsonType)).Post("/withdraw", h.Withdraw)
		})

		r.Route("/", func(r chi.Router) {
			r.Use(middlewares.RequireCookie)

			r.Get("/withdrawals", h.GetWithdrawals)
		})
	})

	return r
}
