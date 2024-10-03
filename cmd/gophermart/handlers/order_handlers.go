package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/storage"
	"net/http"
	"strconv"
	"strings"
	"unicode"
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
	var orderId int64
	if err := decoder.Decode(&orderId); err != nil {
		logger.Log.Info(fmt.Sprintf("decode body error: %s", err))
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if !checkLunAlg(strconv.FormatInt(orderId, 10)) {
		logger.Log.Info("False check Lun Algorithm")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userNameByToken := request.Context().Value(constants.UserNameKey).(string)
	userName := storage.GetUserNameByOrderId(orderId)
	if userName != "" {
		if userName == userNameByToken {
			writer.WriteHeader(http.StatusOK)
			return
		} else if userName != userNameByToken {
			writer.WriteHeader(http.StatusConflict)
			return
		}
	}

	if err := storage.CreateOrder(userNameByToken, orderId); err != nil {
		logger.Log.Info(fmt.Sprintf("Create order error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusAccepted)
}

func ListOrdersWebhook(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Log.Info("Method only accepts GET requests")
		return
	}

	userName := request.Context().Value(constants.UserNameKey).(string)
	orders, err := storage.GetAllOrdersByUserName(userName)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Get all orders error: %s", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	buf := bytes.Buffer{}
	encode := json.NewEncoder(&buf)
	if err := encode.Encode(orders); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(buf.Bytes())
}

// checkLunAlg проверяет корректность номера по алгоритму Луна
func checkLunAlg(number string) bool {
	var sum int
	// Маркер для удвоения каждой второй цифры
	double := false

	// Идем с конца строки к началу
	for i := len(number) - 1; i >= 0; i-- {
		// Получаем текущий символ
		r := rune(number[i])

		// Пропускаем нечисловые символы
		if !unicode.IsDigit(r) {
			continue
		}

		// Преобразуем символ в целое число
		digit := int(r - '0')

		// Если нужно удвоить каждую вторую цифру
		if double {
			digit *= 2
			// Если результат больше 9, вычитаем 9
			if digit > 9 {
				digit -= 9
			}
		}

		// Добавляем к общей сумме
		sum += digit

		// Меняем флаг удвоения для следующей цифры
		double = !double
	}

	// Проверяем, делится ли сумма на 10
	return sum%10 == 0
}
