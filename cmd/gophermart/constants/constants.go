package constants

// contextKey тип для ключа контекста
type contextKey string

// orderStatus статус заказа
type orderStatus string

const (
	// Статусы заказа
	New        orderStatus = "NEW"
	Processing orderStatus = "PROCESSING"
	Invalid    orderStatus = "INVALID"
	Processed  orderStatus = "PROCESSED"

	// UserNameKey ключ для контекста
	UserNameKey contextKey = "userName"
)
