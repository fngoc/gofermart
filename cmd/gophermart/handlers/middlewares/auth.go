package middlewares

import (
	"context"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers/jwt"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"net/http"
)

const CookieName = "token"

// AuthMiddleware — middleware для аунтификации HTTP-запросов.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		cookie, err := r.Cookie(CookieName)

		if err != nil {
			tokenString, err = jwt.BuildJWTString("")

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:  CookieName,
				Value: tokenString,
			})
		} else {
			tokenString = cookie.Value
		}

		userID, err := jwt.GetUserID(tokenString)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			logger.Log.Warn(err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), constants.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
