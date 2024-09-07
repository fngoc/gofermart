package storage

import (
	"context"
	"database/sql"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"time"
)

var Store *sql.DB

// InitializeDB инициализация базы данных
func InitializeDB(dbConf string) error {
	pqx, err := sql.Open("pgx", dbConf)
	if err != nil {
		return err
	}

	Store = pqx

	if err := createTables(pqx); err != nil {
		return err
	}
	return nil
}

func createTables(db *sql.DB) error {
	createTableQuery := ``
	createIndexQuery := ``

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, createTableQuery)
	if err != nil {
		return err
	}
	_, errIdx := db.ExecContext(ctx, createIndexQuery)
	if errIdx != nil {
		return err
	}
	logger.Log.Info("Database table created")
	return nil
}
