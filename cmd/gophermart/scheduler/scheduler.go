package scheduler

import (
	"encoding/json"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/storage"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// AccrualOrderResponse структура для хранения заказа
type AccrualOrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// OrdersForCheck структура для хранения состояния заказов
var OrdersForCheck = map[int]string{}

// Mutex для безопасного доступа к данным заказов
var mutex sync.Mutex

// Канал для передачи обновленных данных заказа
var orderStatusChan = make(chan AccrualOrderResponse)

// Функция для запроса статуса заказа у стороннего сервиса
func requestOrderStatus(orderID int, accrualAddress string) time.Duration {
	var timeOut = 10 * time.Second
	// Выполняем запрос к стороннему сервису
	resp, err := http.Get(fmt.Sprintf("%s/api/orders/%d", accrualAddress, orderID))
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Request error for order %d: %s", orderID, err))
		return timeOut
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var orderResponse AccrualOrderResponse
		err := json.NewDecoder(resp.Body).Decode(&orderResponse)
		if err != nil {
			logger.Log.Warn(fmt.Sprintf("Response decoding error for order %d: %s", orderID, err))
			return timeOut
		}
		// Отправляем обновлённые данные в канал
		orderStatusChan <- orderResponse
		return timeOut
	case http.StatusNoContent:
		logger.Log.Info(fmt.Sprintf("Order %d is not registered in the billing system", orderID))
		return timeOut
	case http.StatusTooManyRequests:
		// Обрабатываем заголовок Retry-After
		retryAfter := resp.Header.Get("Retry-After")
		retrySeconds, err := strconv.Atoi(retryAfter)
		if err != nil {
			log.Printf("Retry-After header parsing error for order %d: %s", orderID, err)
			return timeOut
		}
		logger.Log.Info(fmt.Sprintf("Number of requests for order %d exceeded, repeat in %d seconds", orderID, retrySeconds))
		return time.Duration(retrySeconds) * time.Second
	case http.StatusInternalServerError:
		logger.Log.Warn(fmt.Sprintf("Internal server error for order %d", orderID))
		return timeOut
	default:
		logger.Log.Warn(fmt.Sprintf("Unexpected response code %d for order %d", resp.StatusCode, orderID))
		return timeOut
	}
}

// FetchOrderStatuses горутина для опроса стороннего сервиса
func FetchOrderStatuses(accrualAddress string) {
	for {
		// Проходим по всем заказам
		mutex.Lock()
		var timeOut time.Duration
		for orderID, status := range OrdersForCheck {
			if status == "PROCESSED" || status == "INVALID" {
				delete(OrdersForCheck, orderID)
				continue // Пропускаем заказы с финальными статусами
			}
			timeOut = requestOrderStatus(orderID, accrualAddress) // Для каждого заказа запускаем запрос
		}
		mutex.Unlock()

		time.Sleep(timeOut) // Интервал опроса сервиса
	}
}

// UpdateOrderStatuses горутина для обновления статусов заказов
func UpdateOrderStatuses() {
	for {
		for updatedOrder := range orderStatusChan {
			mutex.Lock()
			orderID, err := strconv.Atoi(updatedOrder.Order)
			if err == nil {
				err := storage.UpdateOrderStatus(orderID, updatedOrder.Status)
				if err != nil {
					logger.Log.Error(fmt.Sprintf("Error updating order status %d, status %s: %s", orderID, updatedOrder.Status, err))
					continue
				}
				logger.Log.Info(fmt.Sprintf("Order status updated %s: %s", updatedOrder.Order, updatedOrder.Status))
			}
			mutex.Unlock()
		}
	}
}
