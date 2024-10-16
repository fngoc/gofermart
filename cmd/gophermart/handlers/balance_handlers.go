package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/handlers/handlermodels"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/storage"
	"github.com/fngoc/gofermart/cmd/gophermart/utils"
	"net/http"
	"strconv"
	"strings"
)

func GetBalanceWebhook(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Method only accepts GET requests")
		return
	}

	userNameByToken := request.Context().Value(constants.UserNameKey).(string)
	userID, err := storage.GetUserIDByName(userNameByToken)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Balance error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	balance, err := storage.GetBalanceByUserID(userID)
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
	var body handlermodels.WithdrawResponse
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

	userNameByToken := request.Context().Value(constants.UserNameKey).(string)
	userID, err := storage.GetUserIDByName(userNameByToken)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Balance withdraw error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !utils.CheckLunAlg(strconv.FormatInt(int64(orderID), 10)) {
		logger.Log.Info("False check Lun Algorithm")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if !storage.IsUserHasOrderID(userID, orderID) {
		logger.Log.Info(fmt.Sprintf("User %d does not have order id %d", userID, orderID))
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	_, err = storage.DeductBalance(userID, orderID, body.Sum)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Deduct balance error: %s", err))
		writer.WriteHeader(http.StatusPaymentRequired)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
