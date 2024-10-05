package storage

import (
	"context"
	"database/sql"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/storage/storage_models"
	"github.com/fngoc/gofermart/cmd/gophermart/utils"
	_ "github.com/jackc/pgx/v5/stdlib"
	"math"
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
		accrual INTEGER,
		status VARCHAR NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`
	createBalancesTableQuery := `
	CREATE TABLE IF NOT EXISTS balances (
	   id SERIAL PRIMARY KEY,
	   user_id INTEGER NOT NULL UNIQUE,
	   current_balance DOUBLE PRECISION,
	   withdrawn INTEGER,
	   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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

	_, err := store.ExecContext(ctx,
		`INSERT INTO users (user_name, password, token) VALUES ($1, $2, $3)`,
		userName, passwordHash, token,
	)
	return err
}

func SetNewTokenByUser(userName, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := store.ExecContext(ctx,
		`UPDATE users SET token = $1 
             	WHERE user_name = $2;`, token, userName)
	return err
}

func GetUserNameByOrderId(orderId int64) string {
	var userName string
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := store.QueryRowContext(ctx,
		`SELECT users.user_name FROM orders
    			JOIN users 
    			    ON orders.user_id = users.id
                WHERE orders.order_id = $1;`, orderId)
	err := row.Scan(&userName)
	if err != nil {
		return ""
	}
	return userName
}

func GetUserIdByName(userName string) (int, error) {
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

func CreateOrder(userID int, orderID int64) error {
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

func GetAllOrdersByUserId(userID int) ([]storage_models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := store.QueryContext(ctx,
		`SELECT order_id, status, accrual, created_at FROM orders
                WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var result []storage_models.Order
	for rows.Next() {
		var orderID string
		var status string
		var accrual sql.NullInt64
		var createdAt string

		if err := rows.Scan(&orderID, &status, &accrual, &createdAt); err != nil {
			return nil, err
		}

		result = append(result, storage_models.Order{
			Number:     orderID,
			Status:     status,
			Accrual:    utils.GetValueByNullInt64(accrual),
			UploadedAt: utils.ConvertTime(createdAt),
		})
	}
	return result, nil
}

func GetBalanceByUserId(userID int) (storage_models.Balance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := store.QueryContext(ctx,
		`SELECT current_balance, withdrawn FROM balances
                WHERE user_id = $1`, userID)
	if err != nil {
		return storage_models.Balance{}, err
	}
	if rows.Err() != nil {
		return storage_models.Balance{}, rows.Err()
	}

	var result storage_models.Balance
	for rows.Next() {
		var currentBalance float64
		var withdrawn int

		if err := rows.Scan(&currentBalance, &withdrawn); err != nil {
			return storage_models.Balance{}, err
		}
		result = storage_models.Balance{
			Current:   math.Round(currentBalance*10) / 10, //преобразование ответа с 1 знаком после запятой
			Withdrawn: withdrawn,
		}
	}

	return result, nil
}

func IsUserHasOrderId(userID, orderID int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists int
	err := store.QueryRowContext(ctx,
		`SELECT 1 FROM orders
         		WHERE order_id = $1 AND user_id = $2`, orderID, userID).Scan(&exists)
	if err != nil {
		return false
	}
	return true
}

func DeductBalance(userID int, amountToDeduct float64) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newBalance float64
	err := store.QueryRowContext(ctx,
		`UPDATE balances
				SET current_balance = current_balance - $1, withdrawn = withdrawn + $1
				WHERE user_id = $2 AND current_balance >= $1 
				RETURNING current_balance
				`, amountToDeduct, userID).Scan(&newBalance)
	if err != nil {
		return 0, err
	}
	return newBalance, nil
}
