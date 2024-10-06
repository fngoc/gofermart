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

func ListWithdrawBalanceWebhook(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Method only accepts GET requests")
		return
	}

	userNameByToken := request.Context().Value(constants.UserNameKey).(string)
	userID, err := storage.GetUserIdByName(userNameByToken)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Transactions error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	transactions, err := storage.GetAllTransactionByUserId(userID)
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
