package main

import (
	"github.com/fngoc/gofermart/cmd/gophermart/configs"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"github.com/fngoc/gofermart/cmd/gophermart/server"
	"github.com/fngoc/gofermart/cmd/gophermart/storage"
)

// main старт программы
func main() {
	if err := logger.Initialize(); err != nil {
		panic(err)
	}

	configs.ParseArgs()

	if configs.HasFlagOrEnvPostgresVariable() {
		if err := storage.InitializeDB(configs.Flags.DBConf); err != nil {
			logger.Log.Fatal(err.Error())
		}
	}

	if err := server.Run(); err != nil {
		logger.Log.Fatal(err.Error())
	}
}
