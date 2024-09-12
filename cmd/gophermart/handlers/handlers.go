package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers/jwt"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers/models"
	"github.com/fngoc/gofermart/cmd/gophermart/hash"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/storage"
	"net/http"
)

// RegisterWebhook обработчик регистрации POST HTTP-запроса
func RegisterWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Register only accepts POST requests")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request models.RegisterRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Log.Info(fmt.Sprintf("Registered user error: %s", err))
		return
	}

	if request.Login == "" || request.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Empty login or password")
		return
	}

	if storage.IsUserCreated(request.Login) {
		w.WriteHeader(http.StatusConflict)
		logger.Log.Info("User already exists")
		return
	}

	passwordHash, err := hash.HashingPassword(request.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Info(fmt.Sprintf("Registered user error: %s", err))
		return
	}

	jwtToken, err := jwt.BuildJWTString(request.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Info(fmt.Sprintf("Registered user error: %s", err))
		return
	}
	if err := storage.CreateUser(request.Login, passwordHash, jwtToken); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Info(fmt.Sprintf("Registered user error: %s", err))
		return
	}
	logger.Log.Info("Registered user successfully")
	w.Header().Set("Authorization", jwtToken)
	w.WriteHeader(http.StatusOK)
}
