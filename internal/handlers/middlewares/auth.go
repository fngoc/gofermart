package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fngoc/gofermart/internal/constants"
	"github.com/fngoc/gofermart/internal/handlers/jwt"
	"github.com/fngoc/gofermart/internal/logger"
)

// AuthMiddleware middleware для аутентификации HTTP-запросов
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token == "" {
			logger.Log.Warn("No auth header found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userName, err := jwt.GetUserNameByToken(token)
		if err != nil {
			logger.Log.Warn(fmt.Sprintf("Decode jwt error: %s", err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), constants.UserNameKey, userName)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
