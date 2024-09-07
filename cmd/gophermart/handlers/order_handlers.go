package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/scheduler"
	"github.com/fngoc/gofermart/cmd/gophermart/storage"
	"net/http"
	"strconv"
	"strings"
)

func LoadOrderWebhook(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Method only accepts POST requests")
		return
	}

	allowedApplicationJSON := strings.Contains(request.Header.Get("Content-Type"), "text/plain")
	if !allowedApplicationJSON {
		logger.Log.Info("need header: 'Content-Type: text/plain'")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(request.Body)
	var orderID int
	if err := decoder.Decode(&orderID); err != nil {
		logger.Log.Info(fmt.Sprintf("decode body error: %s", err))
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := goluhn.Validate(strconv.Itoa(orderID)); err != nil {
		logger.Log.Info("False check Lun Algorithm")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userNameByToken := request.Context().Value(constants.UserNameKey).(string)
	userName := storage.GetUserNameByOrderID(orderID)
	if userName != "" {
		if userName == userNameByToken {
			writer.WriteHeader(http.StatusOK)
			return
		} else if userName != userNameByToken {
			writer.WriteHeader(http.StatusConflict)
			return
		}
	}

	userID, err := storage.GetUserIDByName(userNameByToken)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Create order error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := storage.CreateOrder(userID, orderID); err != nil {
		logger.Log.Info(fmt.Sprintf("Create order error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	scheduler.OrdersForCheck[orderID] = "NEW"
	writer.WriteHeader(http.StatusAccepted)
}

func ListOrdersWebhook(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Method only accepts GET requests")
		return
	}

	userNameByToken := request.Context().Value(constants.UserNameKey).(string)
	userID, err := storage.GetUserIDByName(userNameByToken)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("List order error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	orders, err := storage.GetAllOrdersByUserID(userID)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Get all orders error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		logger.Log.Info("No orders found")
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	buf := bytes.Buffer{}
	encode := json.NewEncoder(&buf)
	if err := encode.Encode(orders); err != nil {
		logger.Log.Warn(fmt.Sprintf("Encode order error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(buf.Bytes())
}
