package handlers

import (
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"net/http"
)

func GetBalanceWebhook(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Method only accepts GET requests")
		return
	}

	//userName := request.Context().Value(constants.UserNameKey).(string)
	//transaction history
}
