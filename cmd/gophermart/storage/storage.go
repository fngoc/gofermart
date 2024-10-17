package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/storage/storagemodels"
	"github.com/fngoc/gofermart/cmd/gophermart/utils"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/shopspring/decimal"
	"time"
)

var store *sql.DB

// InitializeDB инициализация базы данных
func InitializeDB(dbConf string) error {
	pqx, err := sql.Open("pgx", dbConf)
	if err != nil {
		return err
	}

	store = pqx

	if err := createTables(pqx); err != nil {
		return err
	}
	return nil
}

func createTables(db *sql.DB) error {
	createUserTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_name VARCHAR NOT NULL UNIQUE,
		password TEXT NOT NULL,
		token TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	createOrderTableQuery := `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		order_id BIGINT NOT NULL UNIQUE,
		user_id INTEGER NOT NULL,
		accrual NUMERIC(20, 2),
		status VARCHAR NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`
	createBalancesTableQuery := `
	CREATE TABLE IF NOT EXISTS balances (
	   id SERIAL PRIMARY KEY,
	   user_id INTEGER NOT NULL UNIQUE,
	   current_balance NUMERIC(20, 2),
	   withdrawn NUMERIC(20, 2),
	   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	   FOREIGN KEY (user_id) REFERENCES users(id)
	)`

	createTransactionHistoryTableQuery := `
	CREATE TABLE IF NOT EXISTS transaction_history (
	   id SERIAL PRIMARY KEY,
	   user_id INTEGER NOT NULL,
	   order_number BIGINT NOT NULL,
	   transaction_sum NUMERIC(20, 2),
	   processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	   FOREIGN KEY (user_id) REFERENCES users(id)
	)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, errUser := db.ExecContext(ctx, createUserTableQuery)
	if errUser != nil {
		return errUser
	}
	_, errData := db.ExecContext(ctx, createOrderTableQuery)
	if errData != nil {
		return errData
	}
	_, errBalance := db.ExecContext(ctx, createBalancesTableQuery)
	if errBalance != nil {
		return errBalance
	}
	_, errTransaction := db.ExecContext(ctx, createTransactionHistoryTableQuery)
	if errTransaction != nil {
		return errTransaction
	}
	logger.Log.Info("Database table created")
	return nil
}

func IsUserCreated(userName string) bool {
	var isCreated bool
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := store.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM users 
                WHERE user_name = $1)`, userName)
	err := row.Scan(&isCreated)
	if err != nil {
		logger.Log.Error(err.Error())
		return false
	}
	return isCreated
}

func IsUserAuthenticated(userName, passwordHash string) bool {
	var IsAuthenticated bool
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := store.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM users 
                WHERE user_name = $1 AND password = $2)`, userName, passwordHash)
	err := row.Scan(&IsAuthenticated)
	if err != nil {
		return false
	}
	return IsAuthenticated
}

func CreateUser(userName, passwordHash, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := store.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	var userID int
	// Вставляем пользователя и получаем user_id
	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (user_name, password, token) VALUES ($1, $2, $3) 
				RETURNING id`, userName, passwordHash, token).Scan(&userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert user and get user_id: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO balances (user_id, current_balance, withdrawn) VALUES ($1, $2, $3)`,
		userID, 0, 0)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to insert balance: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func SetNewTokenByUser(userName, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := store.ExecContext(ctx,
		`UPDATE users SET token = $1 
             	WHERE user_name = $2;`, token, userName)
	return err
}

func GetUserNameByOrderID(orderID int) string {
	var userName string
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := store.QueryRowContext(ctx,
		`SELECT users.user_name FROM orders
    			JOIN users 
    			    ON orders.user_id = users.id
                WHERE orders.order_id = $1;`, orderID)
	err := row.Scan(&userName)
	if err != nil {
		return ""
	}
	return userName
}

func GetUserIDByName(userName string) (int, error) {
	var id int
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := store.QueryRowContext(ctx,
		`SELECT id FROM users
          		WHERE user_name = $1;`, userName)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func CreateOrder(userID int, orderID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := store.ExecContext(ctx,
		`INSERT INTO orders (user_id, order_id, status) VALUES ($1, $2, $3)`,
		userID, orderID, constants.New)
	if err != nil {
		return err
	}
	return nil
}

func GetAllOrdersByUserID(userID int) ([]storagemodels.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := store.QueryContext(ctx,
		`SELECT order_id, status, accrual, created_at FROM orders
                WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var result []storagemodels.Order
	for rows.Next() {
		var orderID string
		var status string
		var accrual sql.NullString
		var createdAt string

		if err := rows.Scan(&orderID, &status, &accrual, &createdAt); err != nil {
			return nil, err
		}

		var accrualDecimal *decimal.Decimal
		var accrualFloat float64
		if accrual.Valid {
			value, err := decimal.NewFromString(accrual.String)
			if err != nil {
				return nil, err
			}
			if !value.IsZero() {
				accrualDecimal = &value
			}
			f, _ := accrualDecimal.Float64()
			accrualFloat = f
		}

		result = append(result, storagemodels.Order{
			Number:     orderID,
			Status:     status,
			Accrual:    accrualFloat,
			UploadedAt: utils.ConvertTime(createdAt),
		})
	}
	return result, nil
}

func GetBalanceByUserID(userID int) (storagemodels.Balance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := store.QueryContext(ctx,
		`SELECT current_balance, withdrawn FROM balances
                WHERE user_id = $1`, userID)
	if err != nil {
		return storagemodels.Balance{}, err
	}
	if rows.Err() != nil {
		return storagemodels.Balance{}, rows.Err()
	}

	var result storagemodels.Balance
	for rows.Next() {
		var currentBalance string
		var withdrawn string

		if err := rows.Scan(&currentBalance, &withdrawn); err != nil {
			return storagemodels.Balance{}, err
		}

		currentBalanceDecimal, errCurrentBalance := decimal.NewFromString(currentBalance)
		if errCurrentBalance != nil {
			return storagemodels.Balance{}, errCurrentBalance
		}
		currentFloat, _ := currentBalanceDecimal.Float64()

		withdrawnDecimal, errWithdrawn := decimal.NewFromString(withdrawn)
		if errWithdrawn != nil {
			return storagemodels.Balance{}, errWithdrawn
		}
		withdrawnFloat, _ := withdrawnDecimal.Float64()

		result = storagemodels.Balance{
			Current:   currentFloat,
			Withdrawn: withdrawnFloat,
		}
	}

	return result, nil
}

func IsUserHasOrderID(userID, orderID int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var isExist bool
	row := store.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM orders 
                WHERE order_id = $1 AND user_id = $2)`, orderID, userID)
	err := row.Scan(&isExist)
	if err != nil {
		logger.Log.Error(err.Error())
		return false
	}
	return isExist
}

func DeductBalance(userID, orderID int, amountToDeduct float64) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := store.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	var newBalance float64
	err = tx.QueryRowContext(ctx,
		`UPDATE balances
				SET current_balance = current_balance - $1, withdrawn = withdrawn + $1
				WHERE user_id = $2 AND current_balance >= $1
				RETURNING current_balance
				`, amountToDeduct, userID).Scan(&newBalance)
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("failed to update balance: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO transaction_history (user_id, order_number, transaction_sum) VALUES ($1, $2, $3)`,
		userID, orderID, amountToDeduct)
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("failed to insert transaction history: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newBalance, nil
}

func GetAllTransactionByUserID(userID int) ([]storagemodels.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := store.QueryContext(ctx,
		`SELECT order_number, transaction_sum, processed_at FROM transaction_history
                WHERE user_id = $1 ORDER BY processed_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var result []storagemodels.Transaction
	for rows.Next() {
		var orderNumber string
		var transactionSum string
		var processedAt string

		if err := rows.Scan(&orderNumber, &transactionSum, &processedAt); err != nil {
			return nil, err
		}

		transactionSumDecimal, err := decimal.NewFromString(transactionSum)
		if err != nil {
			return nil, err
		}
		transactionSumFloat, _ := transactionSumDecimal.Float64()

		result = append(result, storagemodels.Transaction{
			OrderNumber: orderNumber,
			Sum:         transactionSumFloat,
			ProcessedAt: utils.ConvertTime(processedAt),
		})
	}
	return result, nil
}

func UpdateOrderStatus(orderID int, accrual float64, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := store.ExecContext(ctx,
		`UPDATE orders SET status = $1, accrual = $2 
             	WHERE order_id = $3;`, status, accrual, orderID)
	return err
}
