package storage

import (
	"context"
	"database/sql"
	_ "database/sql"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/storage/storagemodels"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestIsUserCreated тестирует функцию IsUserCreated
func TestIsUserCreated(t *testing.T) {
	// создаем mock базы данных
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// устанавливаем глобальную переменную Store на замоканную базу
	SetDbInstance(SQLStorage{db: db})

	// создаем контекст с таймаутом
	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Замокать запрос проверки существования пользователя
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM users WHERE user_name = \$1\)`).
		WithArgs("testUser").WillReturnRows(rows)

	// Проверяем результат работы функции
	isCreated := Store.IsUserCreated("testUser")
	assert.True(t, isCreated)

	// Проверяем, что все ожидаемые запросы были вызваны
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestIsUserAuthenticated тестирует функцию IsUserAuthenticated
func TestIsUserAuthenticated(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Тест 1: успешная аутентификация
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM users WHERE user_name = \$1 AND password = \$2\)`).
		WithArgs("testUser", "testPasswordHash").
		WillReturnRows(rows)

	isAuthenticated := Store.IsUserAuthenticated("testUser", "testPasswordHash")
	assert.True(t, isAuthenticated)

	// Тест 2: неуспешная аутентификация (неверный логин или пароль)
	rows = sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM users WHERE user_name = \$1 AND password = \$2\)`).
		WithArgs("wrongUser", "wrongPasswordHash").
		WillReturnRows(rows)

	isAuthenticated = Store.IsUserAuthenticated("wrongUser", "wrongPasswordHash")
	assert.False(t, isAuthenticated)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestCreateUser тестирует функцию CreateUser
func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Тест 1: успешное создание пользователя и баланса
	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("testUser", "passwordHash", "token").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectExec(`INSERT INTO balances`).
		WithArgs(1, 0, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = Store.CreateUser("testUser", "passwordHash", "token")
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 2: ошибка при добавлении пользователя
	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("testUser", "passwordHash", "token").
		WillReturnError(fmt.Errorf("insertion error"))

	mock.ExpectRollback()

	err = Store.CreateUser("testUser", "passwordHash", "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert user")

	// Тест 3: ошибка при добавлении баланса
	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("testUser", "passwordHash", "token").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectExec(`INSERT INTO balances`).
		WithArgs(1, 0, 0).
		WillReturnError(fmt.Errorf("balance insertion error"))

	mock.ExpectRollback()

	err = Store.CreateUser("testUser", "passwordHash", "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert balance")

	// Тест 4: ошибка при коммите транзакции
	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("testUser", "passwordHash", "token").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectExec(`INSERT INTO balances`).
		WithArgs(1, 0, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	err = Store.CreateUser("testUser", "passwordHash", "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit transaction")
}

// TestSetNewTokenByUser тестирует функцию SetNewTokenByUser
func TestSetNewTokenByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Тест 1: успешное обновление токена
	mock.ExpectExec(`UPDATE users SET token = \$1 WHERE user_name = \$2`).
		WithArgs("newtoken", "testuser").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = Store.SetNewTokenByUser("testuser", "newtoken")
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 2: ошибка при выполнении запроса обновления токена
	mock.ExpectExec(`UPDATE users SET token = \$1 WHERE user_name = \$2`).
		WithArgs("newtoken", "testuser").
		WillReturnError(sqlmock.ErrCancelled)

	err = Store.SetNewTokenByUser("testuser", "newtoken")
	assert.Error(t, err)
	assert.Equal(t, sqlmock.ErrCancelled, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetUserNameByOrderID тестирует функцию GetUserNameByOrderID
func TestGetUserNameByOrderID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Тест 1: успешное получение имени пользователя
	rows := sqlmock.NewRows([]string{"user_name"}).AddRow("testUser")
	mock.ExpectQuery(`SELECT users.user_name FROM orders JOIN users ON orders.user_id = users.id WHERE orders.order_id = \$1`).
		WithArgs(123).
		WillReturnRows(rows)

	userName := Store.GetUserNameByOrderID(123)
	assert.Equal(t, "testUser", userName)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetUserIDByName тестирует функцию GetUserIDByName
func TestGetUserIDByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Тест 1: успешное получение id пользователя
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`SELECT id FROM users WHERE user_name = \$1`).
		WithArgs("testUser").
		WillReturnRows(rows)

	userID, err := Store.GetUserIDByName("testUser")
	assert.NoError(t, err)
	assert.Equal(t, 1, userID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 2: пользователь не найден
	mock.ExpectQuery(`SELECT id FROM users WHERE user_name = \$1`).
		WithArgs("unknownUser").
		WillReturnError(sql.ErrNoRows)

	userID, err = Store.GetUserIDByName("unknownUser")
	assert.Error(t, err)
	assert.Equal(t, 0, userID)
	assert.Equal(t, sql.ErrNoRows, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 3: ошибка выполнения запроса
	mock.ExpectQuery(`SELECT id FROM users WHERE user_name = \$1`).
		WithArgs("errorUser").
		WillReturnError(sqlmock.ErrCancelled)

	userID, err = Store.GetUserIDByName("errorUser")
	assert.Error(t, err)
	assert.Equal(t, 0, userID)
	assert.Equal(t, sqlmock.ErrCancelled, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestCreateOrder тестирует функцию CreateOrder
func TestCreateOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Тест 1: успешное создание заказа
	mock.ExpectExec(`INSERT INTO orders \(user_id, order_id, status\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(1, 123, constants.New).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = Store.CreateOrder(1, 123)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 2: ошибка при выполнении запроса на создание заказа
	mock.ExpectExec(`INSERT INTO orders \(user_id, order_id, status\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(1, 123, constants.New).
		WillReturnError(sql.ErrConnDone)

	err = Store.CreateOrder(1, 123)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllOrdersByUserID тестирует функцию GetAllOrdersByUserID
func TestGetAllOrdersByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	// Тест 1: успешное получение заказов
	rows := sqlmock.NewRows([]string{"order_id", "status", "accrual", "created_at"}).
		AddRow("123", "NEW", "100.50", "2023-10-22T12:00:00Z").
		AddRow("124", "PROCESSED", sql.NullString{String: "200.75", Valid: true}, "2023-10-21T11:00:00Z")

	mock.ExpectQuery(`SELECT order_id, status, accrual, created_at FROM orders WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs(1).
		WillReturnRows(rows)

	orders, err := Store.GetAllOrdersByUserID(1)
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, "123", orders[0].Number)
	assert.Equal(t, "NEW", orders[0].Status)
	assert.Equal(t, 100.50, orders[0].Accrual)
	assert.Equal(t, "124", orders[1].Number)
	assert.Equal(t, 200.75, orders[1].Accrual)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 2: ошибка при выполнении запроса
	mock.ExpectQuery(`SELECT order_id, status, accrual, created_at FROM orders WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	orders, err = Store.GetAllOrdersByUserID(1)
	assert.Error(t, err)
	assert.Nil(t, orders)
	assert.Equal(t, sql.ErrConnDone, err)

	// Тест 3: ошибка при сканировании строки
	rowsWithScanError := sqlmock.NewRows([]string{"order_id", "status", "accrual", "created_at"}).
		AddRow(nil, nil, nil, nil) // Неправильные данные

	mock.ExpectQuery(`SELECT order_id, status, accrual, created_at FROM orders WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs(1).
		WillReturnRows(rowsWithScanError)

	orders, err = Store.GetAllOrdersByUserID(1)
	assert.Error(t, err)
	assert.Nil(t, orders)
}

// TestGetBalanceByUserID тестирует функцию GetBalanceByUserID
func TestGetBalanceByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	// Тест 1: успешное получение баланса пользователя
	rows := sqlmock.NewRows([]string{"current_balance", "withdrawn"}).
		AddRow("100.50", "50.25")

	mock.ExpectQuery(`SELECT current_balance, withdrawn FROM balances WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(rows)

	balance, err := Store.GetBalanceByUserID(1)
	assert.NoError(t, err)
	assert.Equal(t, 100.50, balance.Current)
	assert.Equal(t, 50.25, balance.Withdrawn)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 2: ошибка при выполнении запроса
	mock.ExpectQuery(`SELECT current_balance, withdrawn FROM balances WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	balance, err = Store.GetBalanceByUserID(1)
	assert.Error(t, err)
	assert.Equal(t, storagemodels.Balance{}, balance)
	assert.Equal(t, sql.ErrConnDone, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 3: ошибка при сканировании данных из строки
	rowsWithScanError := sqlmock.NewRows([]string{"current_balance", "withdrawn"}).
		AddRow("invalid_balance", "50.25") // Неверный формат баланса

	mock.ExpectQuery(`SELECT current_balance, withdrawn FROM balances WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(rowsWithScanError)

	balance, err = Store.GetBalanceByUserID(1)
	assert.Error(t, err)
	assert.Equal(t, storagemodels.Balance{}, balance)
}

// TestDeductBalance тестирует функцию DeductBalance
func TestDeductBalance(t *testing.T) {
	// создаем mock базы данных
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Тест 1: успешное вычитание баланса и добавление в историю транзакций
	mock.ExpectBegin()

	mock.ExpectQuery(`UPDATE balances SET current_balance = current_balance - \$1, withdrawn = withdrawn + \$1 WHERE user_id = \$2 AND current_balance >= \$1 RETURNING current_balance`).
		WithArgs(100.0, 1).
		WillReturnRows(sqlmock.NewRows([]string{"current_balance"}).AddRow(900.0))

	mock.ExpectExec(`INSERT INTO transaction_history \(user_id, order_number, transaction_sum\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(1, 123, 100.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	newBalance, err := Store.DeductBalance(1, 123, 100.0)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 2: ошибка при начале транзакции
	mock.ExpectBegin().WillReturnError(fmt.Errorf("transaction begin error"))

	newBalance, err = Store.DeductBalance(1, 123, 100.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
	assert.Equal(t, 0.0, newBalance)

	// Тест 3: ошибка при обновлении баланса
	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE balances SET current_balance = current_balance - \$1, withdrawn = withdrawn + \$1 WHERE user_id = \$2 AND current_balance >= \$1 RETURNING current_balance`).
		WithArgs(100.0, 1).
		WillReturnError(fmt.Errorf("update balance error"))
	mock.ExpectRollback()

	newBalance, err = Store.DeductBalance(1, 123, 100.0)
	assert.Error(t, err)
	assert.Equal(t, 0.0, newBalance)

	// Тест 4: ошибка при добавлении в историю транзакций
	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE balances SET current_balance = current_balance - \$1, withdrawn = withdrawn + \$1 WHERE user_id = \$2 AND current_balance >= \$1 RETURNING current_balance`).
		WithArgs(100.0, 1).
		WillReturnRows(sqlmock.NewRows([]string{"current_balance"}).AddRow(900.0))

	mock.ExpectExec(`INSERT INTO transaction_history \(user_id, order_number, transaction_sum\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(1, 123, 100.0).
		WillReturnError(fmt.Errorf("insert history error"))

	mock.ExpectRollback()

	newBalance, err = Store.DeductBalance(1, 123, 100.0)
	assert.Error(t, err)
	assert.Equal(t, 0.0, newBalance)

	// Тест 5: ошибка при коммите транзакции
	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE balances SET current_balance = current_balance - \$1, withdrawn = withdrawn + \$1 WHERE user_id = \$2 AND current_balance >= \$1 RETURNING current_balance`).
		WithArgs(100.0, 1).
		WillReturnRows(sqlmock.NewRows([]string{"current_balance"}).AddRow(900.0))

	mock.ExpectExec(`INSERT INTO transaction_history \(user_id, order_number, transaction_sum\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(1, 123, 100.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	newBalance, err = Store.DeductBalance(1, 123, 100.0)
	assert.Error(t, err)
	assert.Equal(t, 0.0, newBalance)
}

// TestGetAllTransactionByUserID тестирует функцию GetAllTransactionByUserID
func TestGetAllTransactionByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	// Тест 1: успешное получение истории транзакций пользователя
	rows := sqlmock.NewRows([]string{"order_number", "transaction_sum", "processed_at"}).
		AddRow("123", "100.50", "2023-10-22T12:00:00Z").
		AddRow("124", "200.75", "2023-10-21T11:00:00Z")

	mock.ExpectQuery(`SELECT order_number, transaction_sum, processed_at FROM transaction_history WHERE user_id = \$1 ORDER BY processed_at DESC`).
		WithArgs(1).
		WillReturnRows(rows)

	transactions, err := Store.GetAllTransactionByUserID(1)
	assert.NoError(t, err)
	assert.Len(t, transactions, 2)
	assert.Equal(t, "123", transactions[0].OrderNumber)
	assert.Equal(t, 100.50, transactions[0].Sum)
	assert.Equal(t, "124", transactions[1].OrderNumber)
	assert.Equal(t, 200.75, transactions[1].Sum)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Тест 2: ошибка при выполнении запроса
	mock.ExpectQuery(`SELECT order_number, transaction_sum, processed_at FROM transaction_history WHERE user_id = \$1 ORDER BY processed_at DESC`).
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	transactions, err = Store.GetAllTransactionByUserID(1)
	assert.Error(t, err)
	assert.Nil(t, transactions)
	assert.Equal(t, sql.ErrConnDone, err)

	// Тест 3: ошибка при обработке строк (rows.Err)
	rowsWithError := sqlmock.NewRows([]string{"order_number", "transaction_sum", "processed_at"}).
		AddRow("125", "300.50", "2023-10-20T10:00:00Z").
		RowError(0, sql.ErrNoRows)

	mock.ExpectQuery(`SELECT order_number, transaction_sum, processed_at FROM transaction_history WHERE user_id = \$1 ORDER BY processed_at DESC`).
		WithArgs(1).
		WillReturnRows(rowsWithError)

	transactions, err = Store.GetAllTransactionByUserID(1)
	assert.Nil(t, transactions)

	// Тест 4: ошибка при сканировании строки
	rowsWithScanError := sqlmock.NewRows([]string{"order_number", "transaction_sum", "processed_at"}).
		AddRow("invalid_order", "invalid_sum", "2023-10-22T12:00:00Z") // Неправильные данные

	mock.ExpectQuery(`SELECT order_number, transaction_sum, processed_at FROM transaction_history WHERE user_id = \$1 ORDER BY processed_at DESC`).
		WithArgs(1).
		WillReturnRows(rowsWithScanError)

	transactions, err = Store.GetAllTransactionByUserID(1)
	assert.Error(t, err)
	assert.Nil(t, transactions)
}

// TestUpdateAccrualData тестирует функцию UpdateAccrualData
func TestUpdateAccrualData(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	SetDbInstance(SQLStorage{db: db})

	// Тест 1: успешное обновление данных заказа и баланса
	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT user_id FROM orders WHERE order_id = \$1`).
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	mock.ExpectExec(`UPDATE orders SET status = \$1, accrual = \$2 WHERE order_id = \$3`).
		WithArgs("PROCESSED", 100.0, 123).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE balances SET current_balance = current_balance + \$1 WHERE user_id = \$2`).
		WithArgs(100.0, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = Store.UpdateAccrualData(123, 100.0, "PROCESSED")

	err = mock.ExpectationsWereMet()

	// Тест 2: ошибка при начале транзакции
	mock.ExpectBegin().WillReturnError(fmt.Errorf("transaction begin error"))

	err = Store.UpdateAccrualData(123, 100.0, "PROCESSED")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")

	// Тест 3: ошибка при поиске userID по orderID
	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT user_id FROM orders WHERE order_id = \$1`).
		WithArgs(123).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	err = Store.UpdateAccrualData(123, 100.0, "PROCESSED")
	assert.Error(t, err)

	// Тест 4: ошибка при обновлении заказа
	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT user_id FROM orders WHERE order_id = \$1`).
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	mock.ExpectExec(`UPDATE orders SET status = \$1, accrual = \$2 WHERE order_id = \$3`).
		WithArgs("PROCESSED", 100.0, 123).
		WillReturnError(fmt.Errorf("update order error"))

	mock.ExpectRollback()

	err = Store.UpdateAccrualData(123, 100.0, "PROCESSED")
	assert.Error(t, err)

	// Тест 5: ошибка при обновлении баланса
	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT user_id FROM orders WHERE order_id = \$1`).
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	mock.ExpectExec(`UPDATE orders SET status = \$1, accrual = \$2 WHERE order_id = \$3`).
		WithArgs("PROCESSED", 100.0, 123).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE balances SET current_balance = current_balance + \$1 WHERE user_id = \$2`).
		WithArgs(100.0, 1).
		WillReturnError(fmt.Errorf("update balance error"))

	mock.ExpectRollback()

	err = Store.UpdateAccrualData(123, 100.0, "PROCESSED")
	assert.Error(t, err)

	// Тест 6: ошибка при коммите транзакции
	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT user_id FROM orders WHERE order_id = \$1`).
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	mock.ExpectExec(`UPDATE orders SET status = \$1, accrual = \$2 WHERE order_id = \$3`).
		WithArgs("PROCESSED", 100.0, 123).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE balances SET current_balance = current_balance + \$1 WHERE user_id = \$2`).
		WithArgs(100.0, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	err = Store.UpdateAccrualData(123, 100.0, "PROCESSED")
	assert.Error(t, err)
}
