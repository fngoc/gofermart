package server

import (
	"github.com/fngoc/gofermart/cmd/gophermart/configs"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers/middlewares"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/scheduler"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// Run запуск сервера
func Run() error {
	logger.Log.Info("Starting server")

	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		//auth
		r.Post("/register", logger.RequestLogger(middlewares.GzipMiddleware(handlers.RegisterWebhook)))
		r.Post("/login", logger.RequestLogger(middlewares.GzipMiddleware(handlers.AuntificationWebhook)))

		//order
		r.Post("/orders", logger.RequestLogger(middlewares.AuthMiddleware(middlewares.GzipMiddleware(handlers.LoadOrderWebhook))))
		r.Get("/orders", logger.RequestLogger(middlewares.AuthMiddleware(middlewares.GzipMiddleware(handlers.ListOrdersWebhook))))

		//balance
		r.Get("/balance", logger.RequestLogger(middlewares.AuthMiddleware(middlewares.GzipMiddleware(handlers.GetBalanceWebhook))))
		r.Post("/balance/withdraw", logger.RequestLogger(middlewares.AuthMiddleware(middlewares.GzipMiddleware(handlers.PostWithdrawBalanceWebhook))))

		//withdrawals
		r.Get("/withdrawals", logger.RequestLogger(middlewares.AuthMiddleware(middlewares.GzipMiddleware(handlers.ListWithdrawalsBalanceWebhook))))
	})

	go scheduler.FetchOrderStatuses(configs.Flags.AccrualAddress)
	go scheduler.UpdateOrderStatuses()

	return http.ListenAndServe(configs.Flags.ServerAddress, r)
}
