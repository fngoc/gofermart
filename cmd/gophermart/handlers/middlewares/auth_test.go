package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers/jwt"
	"github.com/stretchr/testify/assert"
)

// mockHandler фейковый обработчик для проверки вызова следующего обработчика
func mockHandler(w http.ResponseWriter, r *http.Request) {
	userName, ok := r.Context().Value(constants.UserNameKey).(string)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Write([]byte("Hello, " + userName))
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No Authorization header",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "Valid token",
			token:          "%s", // Заменим на валидный токен
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, testUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Генерируем валидный токен для успешного случая
			if strings.Contains(tt.token, "%s") {
				tokenString, err := jwt.BuildJWTByUserName("testUser")
				assert.NoError(t, err)
				tt.token = strings.Replace(tt.token, "%s", tokenString, 1)
			}

			// Создаем запрос с токеном
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			// Записываем результат ответа
			rr := httptest.NewRecorder()

			// Применяем middleware и запускаем тестовый обработчик
			handler := AuthMiddleware(mockHandler)
			handler.ServeHTTP(rr, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Проверяем тело ответа для валидного токена
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
