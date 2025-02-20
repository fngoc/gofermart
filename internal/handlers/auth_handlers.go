package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fngoc/gofermart/internal/handlers/handlermodels"
	"github.com/fngoc/gofermart/internal/handlers/jwt"
	"github.com/fngoc/gofermart/internal/hash"
	"github.com/fngoc/gofermart/internal/logger"
	"github.com/fngoc/gofermart/internal/storage"
)

// RegisterWebhook обработчик регистрации, POST HTTP-запрос
func RegisterWebhook(writer http.ResponseWriter, request *http.Request) {
	body, err := authCheckRequest(request)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info(fmt.Sprintf("Registered check request error: %s", err))
		return
	}

	if storage.Store.IsUserCreated(body.Login) {
		writer.WriteHeader(http.StatusConflict)
		logger.Log.Info("User already exists")
		return
	}

	passwordHash, err := hash.HashingPassword(body.Password)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Log.Warn(fmt.Sprintf("Registered user error: %s", err))
		return
	}

	jwtToken, err := jwt.BuildJWTByUserName(body.Login)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Log.Warn(fmt.Sprintf("Registered user error: %s", err))
		return
	}
	if err := storage.Store.CreateUser(body.Login, passwordHash, jwtToken); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Log.Warn(fmt.Sprintf("Registered user error: %s", err))
		return
	}

	logger.Log.Info(fmt.Sprintf("Registered user '%s' successfully", body.Login))
	writer.Header().Set("Authorization", jwtToken)
	writer.WriteHeader(http.StatusOK)
}

// AuntificationWebhook обработчик аутентификации, POST HTTP-запрос
func AuntificationWebhook(writer http.ResponseWriter, request *http.Request) {
	body, err := authCheckRequest(request)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info(fmt.Sprintf("Auntification check request error: %s", err))
		return
	}

	passwordHash, err := hash.HashingPassword(body.Password)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Log.Warn(fmt.Sprintf("Auntification user error: %s", err))
		return
	}

	if !storage.Store.IsUserAuthenticated(body.Login, passwordHash) {
		writer.WriteHeader(http.StatusUnauthorized)
		logger.Log.Info("Bad username or password")
		return
	}

	jwtToken, err := jwt.BuildJWTByUserName(body.Login)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Log.Warn(fmt.Sprintf("Auntification user error: %s", err))
		return
	}

	if err := storage.Store.SetNewTokenByUser(body.Login, jwtToken); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Log.Warn(fmt.Sprintf("Auntification user error: %s", err))
		return
	}

	logger.Log.Info(fmt.Sprintf("Login user '%s' successfully", body.Login))
	writer.Header().Set("Authorization", jwtToken)
	writer.WriteHeader(http.StatusOK)
}

// authCheckRequest общие проверки для HTTP-запросов
func authCheckRequest(request *http.Request) (handlermodels.AuthRequest, error) {
	if request.Method != http.MethodPost {
		return handlermodels.AuthRequest{}, fmt.Errorf("method only accepts POST requests")
	}

	allowedApplicationJSON := strings.Contains(request.Header.Get("Content-Type"), "application/json")
	if !allowedApplicationJSON {
		return handlermodels.AuthRequest{}, fmt.Errorf("need header: 'Content-Type: application/json'")
	}

	decoder := json.NewDecoder(request.Body)
	var body handlermodels.AuthRequest
	if err := decoder.Decode(&body); err != nil {
		return handlermodels.AuthRequest{}, fmt.Errorf("decode body error: %s", err)
	}

	if body.Login == "" || body.Password == "" {
		return handlermodels.AuthRequest{}, fmt.Errorf("empty login or password")
	}

	return body, nil
}
