package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/storage"
	"net/http"
)

// ListWithdrawalsBalanceWebhook получение истории операций, GET HTTP-запрос
func ListWithdrawalsBalanceWebhook(writer http.ResponseWriter, request *http.Request) {
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
		logger.Log.Info(fmt.Sprintf("Transactions error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	transactions, err := storage.Store.GetAllTransactionByUserID(userID)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Transactions error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(transactions) == 0 {
		logger.Log.Info("No transactions found")
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	buf := bytes.Buffer{}
	encode := json.NewEncoder(&buf)
	if err := encode.Encode(transactions); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(buf.Bytes())
}
