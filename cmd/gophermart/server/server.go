package server

import (
	"github.com/fngoc/gofermart/cmd/gophermart/configs"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers/middlewares"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// Run запуск сервера
func Run() error {
	logger.Log.Info("Starting server")

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", logger.RequestLogger(middlewares.AuthMiddleware(middlewares.GzipMiddleware(handlers.TestWebhook))))
	})

	return http.ListenAndServe(configs.Flags.ServerAddress, r)
}
