package storage

import (
	"context"
	"database/sql"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
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
	createDataTableQuery := ``
	createUserTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		uuid SERIAL PRIMARY KEY,
		user_name VARCHAR NOT NULL UNIQUE,
		password TEXT NOT NULL,
		token TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	createIndexQuery := ``

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, errData := db.ExecContext(ctx, createDataTableQuery)
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

	row := store.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE user_name = $1)`, userName)
	err := row.Scan(&isCreated)
	if err != nil {
		logger.Log.Error(err.Error())
		return false
	}
	return isCreated
}

func CreateUser(userName string, passwordHash string, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := store.ExecContext(ctx, `INSERT INTO users (user_name, password, token) VALUES ($1, $2, $3)`,
		userName, passwordHash, token,
	)
	return err
}
