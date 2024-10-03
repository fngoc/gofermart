package storage

import (
	"context"
	"database/sql"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/storage/storage_models"
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
	createOrderTableQuery := `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		order_id BIGINT NOT NULL UNIQUE,
		user_name VARCHAR NOT NULL,
		accrual INTEGER,
		status VARCHAR NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	createUserTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_name VARCHAR NOT NULL UNIQUE,
		password TEXT NOT NULL,
		token TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	createIndexQuery := ``

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, errData := db.ExecContext(ctx, createOrderTableQuery)
	if errData != nil {
		return errData
	}
	_, errUser := db.ExecContext(ctx, createUserTableQuery)
	if errUser != nil {
		return errUser
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
		`SELECT EXISTS (SELECT 1 FROM users WHERE user_name = $1)`, userName)
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
		`SELECT EXISTS (SELECT 1 FROM users WHERE user_name = $1 AND password = $2)`, userName, passwordHash)
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
		`UPDATE users SET token = $1 WHERE user_name = $2;`, token, userName)
	return err
}

func GetUserNameByOrderId(orderId int64) string {
	var userName string
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := store.QueryRowContext(ctx,
		`SELECT user_name FROM orders WHERE order_id = $1`, orderId)
	err := row.Scan(&userName)
	if err != nil {
		return ""
	}
	return userName
}

func CreateOrder(userName string, orderID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := store.ExecContext(ctx,
		`INSERT INTO orders (user_name, order_id, status) VALUES ($1, $2, $3)`, userName, orderID, "NEW")
	if err != nil {
		return err
	}
	return nil
}

func GetAllOrdersByUserName(userName string) ([]storage_models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := store.QueryContext(ctx,
		`SELECT order_id, status, accrual, created_at FROM orders WHERE user_name = $1 ORDER BY created_at`, userName)
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
			Accrual:    getValueByNullInt64(accrual),
			UploadedAt: convertTime(createdAt),
		})
	}
	return result, nil
}

func getValueByNullInt64(value sql.NullInt64) int {
	if value.Valid {
		return int(value.Int64)
	}
	return 0
}

func convertTime(t string) string {
	parsedTime, err := time.Parse(time.RFC3339Nano, t)
	if err != nil {
		logger.Log.Error(err.Error())
		return t
	}
	location := time.FixedZone("UTC+3", 3*60*60)
	timeInZone := parsedTime.In(location)
	formattedTime := timeInZone.Format(time.RFC3339)
	return formattedTime
}
