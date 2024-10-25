package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestBuildJWTByUserName(t *testing.T) {
	tests := []struct {
		name     string
		userName string
	}{
		{
			name:     "Valid token creation",
			userName: "testUser",
		},
		{
			name:     "Token creation with empty username",
			userName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Генерация токена
			tokenString, err := BuildJWTByUserName(tt.userName)
			assert.NoError(t, err)
			assert.NotEmpty(t, tokenString)

			// Проверка, что токен действителен и содержит правильное имя пользователя
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims,
				func(token *jwt.Token) (interface{}, error) {
					return []byte(secretKey), nil
				})

			assert.NoError(t, err)
			assert.True(t, token.Valid)
			assert.Equal(t, tt.userName, claims.UserName)
			assert.WithinDuration(t, time.Now().Add(tokenExp), claims.ExpiresAt.Time, time.Minute)
		})
	}
}

func TestGetUserNameByToken(t *testing.T) {
	tests := []struct {
		name        string
		userName    string
		expectError bool
	}{
		{
			name:        "Valid token",
			userName:    "testUser",
			expectError: false,
		},
		{
			name:        "Invalid token",
			userName:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokenString string

			// Создаём токен, если это валидный случай
			if !tt.expectError {
				var err error
				tokenString, err = BuildJWTByUserName(tt.userName)
				assert.NoError(t, err)
			} else {
				// Некорректный токен
				tokenString = "invalid.token.string"
			}

			// Тестируем получение имени пользователя из токена
			userName, err := GetUserNameByToken(tokenString)
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, userName)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userName, userName)
			}
		})
	}
}
