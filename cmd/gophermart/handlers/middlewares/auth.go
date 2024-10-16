package middlewares

import (
	"context"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers/jwt"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"net/http"
)

// cookieName наименование куки
const cookieName = "token"

// AuthMiddleware — middleware для аунтификации HTTP-запросов.
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
