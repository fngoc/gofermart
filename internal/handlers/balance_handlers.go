package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/fngoc/gofermart/internal/constants"
	"github.com/fngoc/gofermart/internal/handlers/handlermodels"
	"github.com/fngoc/gofermart/internal/logger"
	"github.com/fngoc/gofermart/internal/storage"
)

// GetBalanceWebhook обработчик получения баланса, GET HTTP-запрос
func GetBalanceWebhook(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Method only accepts GET requests")
		return
	}

	userNameFromToken, ok := request.Context().Value(constants.UserNameKey).(string)
	if !ok {
		logger.Log.Warn("Something went wrong with jwt token")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, err := storage.Store.GetUserIDByName(userNameFromToken)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Balance error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	balance, err := storage.Store.GetBalanceByUserID(userID)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Balance error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf := bytes.Buffer{}
	encode := json.NewEncoder(&buf)
	if err := encode.Encode(balance); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(buf.Bytes())
}

// PostWithdrawBalanceWebhook обработчик списания баланса, POST HTTP-запрос
func PostWithdrawBalanceWebhook(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Method only accepts POST requests")
		return
	}

	allowedApplicationJSON := strings.Contains(request.Header.Get("Content-Type"), "application/json")
	if !allowedApplicationJSON {
		logger.Log.Info("Need header: 'Content-Type: application/json'")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(request.Body)
	var body handlermodels.WithdrawRequest
	if err := decoder.Decode(&body); err != nil {
		logger.Log.Info(fmt.Sprintf("Decode body error: %s", err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	orderID, err := strconv.Atoi(body.Order)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Order error: %s", err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	userNameFromToken, ok := request.Context().Value(constants.UserNameKey).(string)
	if !ok {
		logger.Log.Warn("Something went wrong with jwt token")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, err := storage.Store.GetUserIDByName(userNameFromToken)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Balance withdraw error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := goluhn.Validate(body.Order); err != nil {
		logger.Log.Info("False check Lun Algorithm")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	_, err = storage.Store.DeductBalance(userID, orderID, body.Sum)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Deduct balance error: %s", err))
		writer.WriteHeader(http.StatusPaymentRequired)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
