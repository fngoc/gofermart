package constants

// contextKey тип для ключа контекста
type contextKey string

const (
	// New статус нового заказа
	New string = "NEW"
	// Processed статус завершенного заказа
	Processed string = "PROCESSED"
	// Invalid статус не законченного заказа
	Invalid string = "INVALID"

	// UserNameKey ключ для контекста
	UserNameKey contextKey = "userName"
)
