package storage

import (
	"context"
	"database/sql"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/storage/storage_models"
	"github.com/fngoc/gofermart/cmd/gophermart/utils"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	   balance INTEGER,
	   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	   FOREIGN KEY (user_id) REFERENCES users(id)
	)`

	createIndexQuery := ``

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
	_, errIdx := db.ExecContext(ctx, createIndexQuery)
	if errIdx != nil {
		return errIdx
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

func GetAllOrdersByUserName(userID int) ([]storage_models.Order, error) {
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
