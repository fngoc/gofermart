package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fngoc/gofermart/internal/constants"
	"github.com/fngoc/gofermart/internal/logger"
	"github.com/fngoc/gofermart/internal/storage"
)

// AccrualOrderResponse структура ответа
type AccrualOrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// OrderManager структура менеджера заказов
type OrderManager struct {
	// ordersForCheck структура для хранения состояния заказов
	ordersForCheck map[int]string
	// mutex для безопасного доступа к данным заказов
	mutex sync.RWMutex
	// orderStatusChan канал для передачи обновленных данных заказа
	orderStatusChan chan AccrualOrderResponse
}

// orderManagerInstant инстанс менеджера заказов
var orderManagerInstant *OrderManager

// once выполнится ровно один раз
var once sync.Once

// LazyInitialiseOrderManager ленивая инициализация orderManagerInstant,
// в момент обращения к очереди, во время тестов или во время опроса сервиса accrual
func LazyInitialiseOrderManager() {
	once.Do(func() {
		orderManagerInstant = &OrderManager{
			ordersForCheck:  make(map[int]string),
			mutex:           sync.RWMutex{},
			orderStatusChan: make(chan AccrualOrderResponse),
		}
	})
}

// AddOrderInQueue добавление заказа в очередь на обновление
func AddOrderInQueue(orderID int) {
	LazyInitialiseOrderManager()
	orderManagerInstant.mutex.Lock()
	orderManagerInstant.ordersForCheck[orderID] = constants.New
	orderManagerInstant.mutex.Unlock()
}

// requestOrderStatus функция для запроса статуса заказа у стороннего сервиса
func requestOrderStatus(orderID int, accrualAddress string) time.Duration {
	var timeOut = 2 * time.Second
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
		orderManagerInstant.orderStatusChan <- orderResponse
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
	LazyInitialiseOrderManager()
	for {
		// Проходим по всем заказам
		orderManagerInstant.mutex.RLock()
		var timeOut time.Duration
		for orderID, status := range orderManagerInstant.ordersForCheck {
			if status == constants.Processed || status == constants.Invalid {
				delete(orderManagerInstant.ordersForCheck, orderID)
				continue // Пропускаем заказы с финальными статусами
			}
			timeOut = requestOrderStatus(orderID, accrualAddress) // Для каждого заказа запускаем запрос
		}
		orderManagerInstant.mutex.RUnlock()

		time.Sleep(timeOut) // Интервал опроса сервиса
	}
}

// UpdateOrderStatuses горутина для обновления статусов заказов
func UpdateOrderStatuses() {
	LazyInitialiseOrderManager()
	defer close(orderManagerInstant.orderStatusChan)
	for {
		for updatedOrder := range orderManagerInstant.orderStatusChan {
			orderManagerInstant.mutex.Lock()
			orderID, err := strconv.Atoi(updatedOrder.Order)
			if err == nil {
				err := storage.Store.UpdateAccrualData(orderID, updatedOrder.Accrual, updatedOrder.Status)
				if err != nil {
					logger.Log.Error(fmt.Sprintf("Error updating order status %d, status %s: %s", orderID, updatedOrder.Status, err))
					continue
				}
				logger.Log.Info(fmt.Sprintf("Order status updated %s: %s", updatedOrder.Order, updatedOrder.Status))
			}
			orderManagerInstant.mutex.Unlock()
		}
	}
}
