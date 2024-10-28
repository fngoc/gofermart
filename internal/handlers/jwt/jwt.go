package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserName
type Claims struct {
	jwt.RegisteredClaims
	UserName string
}

const (
	// tokenExp жизнь токена
	tokenExp = time.Hour * 3
	// secretKey секрет
	secretKey = "super-secret-key"
)

// BuildJWTByUserName создаёт токен и возвращает его в виде строки.
func BuildJWTByUserName(userName string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		// собственное утверждение
		UserName: userName,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

// GetUserNameByToken получить имя пользователя из токена
func GetUserNameByToken(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	return claims.UserName, nil
}
