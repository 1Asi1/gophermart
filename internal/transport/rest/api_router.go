package rest

import (
	"github.com/1Asi1/gophermart/internal/service"
	"github.com/1Asi1/gophermart/internal/transport/rest/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type APIRouter struct {
	*chi.Mux
}

func New(s service.Service, log zerolog.Logger) APIRouter {
	router := chi.NewRouter()
	h := newHandlers(s, log)

	router.Use(middleware.DefaultLogger)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
		r.Post("/orders", middlewares.Authorization(h.createOrder, s))
		r.Get("/orders", middlewares.Authorization(h.getOrders, s))
		r.Get("/balance", middlewares.Authorization(h.getBalance, s))
		r.Post("/balance/withdraw", middlewares.Authorization(h.withdraw, s))
		r.Get("/withdrawals", middlewares.Authorization(h.getWithdrawals, s))
	})

	return APIRouter{Mux: router}
}
