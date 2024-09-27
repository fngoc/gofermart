package middlewares

import (
	"context"
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
		var userName string
		cookie, _ := r.Cookie(cookieName)
		token := r.Header.Get("Authorization")

		if token == "" && cookie == nil {
			logger.Log.Warn("No auth cookie and header found")
			w.WriteHeader(http.StatusUnauthorized)
			next.ServeHTTP(w, r)
			return
		} else {
			if cookie != nil {
				token = cookie.Value
			}
			userNameString, err := jwt.GetUserNameByToken(token)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				logger.Log.Warn(err.Error())
				return
			}
			userName = userNameString
		}

		ctx := context.WithValue(r.Context(), constants.UserNameKey, userName)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
