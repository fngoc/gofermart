package handlers

import "github.com/fngoc/gofermart/cmd/gophermart/storage/storagemodels"

// mockStorage имитация хранилища для тестов
type mockStorage struct {
	IsUserCreatedFunc             func(userName string) bool
	IsUserAuthenticatedFunc       func(userName, passwordHash string) bool
	CreateUserFunc                func(userName, passwordHash, token string) error
	GetUserIDByNameFunc           func(userName string) (int, error)
	GetAllTransactionByUserIDFunc func(userID int) ([]storagemodels.Transaction, error)
	SetNewTokenByUserFunc         func(userName, token string) error
	GetUserNameByOrderIDFunc      func(orderID int) string
	CreateOrderFunc               func(userID int, orderID int) error
	GetAllOrdersByUserIDFunc      func(userID int) ([]storagemodels.Order, error)
	GetBalanceByUserIDFunc        func(userID int) (storagemodels.Balance, error)
	DeductBalanceFunc             func(userID, orderID int, amountToDeduct float64) (float64, error)
	UpdateAccrualDataFunc         func(orderID int, accrual float64, status string) error
}

func (m *mockStorage) IsUserCreated(userName string) bool {
	return m.IsUserCreatedFunc(userName)
}

func (m *mockStorage) IsUserAuthenticated(userName, passwordHash string) bool {
	return m.IsUserAuthenticatedFunc(userName, passwordHash)
}

func (m *mockStorage) CreateUser(userName, passwordHash, token string) error {
	return m.CreateUserFunc(userName, passwordHash, token)
}

func (m *mockStorage) SetNewTokenByUser(userName, token string) error {
	return m.SetNewTokenByUserFunc(userName, token)
}

func (m *mockStorage) GetUserNameByOrderID(orderID int) string {
	return m.GetUserNameByOrderIDFunc(orderID)
}

func (m *mockStorage) CreateOrder(userID int, orderID int) error {
	return m.CreateOrderFunc(userID, orderID)
}

func (m *mockStorage) GetAllOrdersByUserID(userID int) ([]storagemodels.Order, error) {
	return m.GetAllOrdersByUserIDFunc(userID)
}

func (m *mockStorage) GetBalanceByUserID(userID int) (storagemodels.Balance, error) {
	return m.GetBalanceByUserIDFunc(userID)
}

func (m *mockStorage) DeductBalance(userID, orderID int, amountToDeduct float64) (float64, error) {
	return m.DeductBalanceFunc(userID, orderID, amountToDeduct)
}

func (m *mockStorage) UpdateAccrualData(orderID int, accrual float64, status string) error {
	return m.UpdateAccrualDataFunc(orderID, accrual, status)
}

func (m *mockStorage) GetUserIDByName(userName string) (int, error) {
	return m.GetUserIDByNameFunc(userName)
}

func (m *mockStorage) GetAllTransactionByUserID(userID int) ([]storagemodels.Transaction, error) {
	return m.GetAllTransactionByUserIDFunc(userID)
}
