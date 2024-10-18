package constants

// contextKey тип для ключа контекста
type contextKey string

// orderStatus статус заказа
type orderStatus string

const (
	// New статусы заказа
	New orderStatus = "NEW"

	// UserNameKey ключ для контекста
	UserNameKey contextKey = "userName"
)
