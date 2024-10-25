package configs

import (
	"flag"
	"fmt"
	"os"

	"github.com/fngoc/gofermart/internal/logger"
)

// flags тип аргументы программы
type flags struct {
	AccrualAddress string
	ServerAddress  string
	DBConf         string
}

const (
	defaultServerAddress  string = "localhost:8080"
	defaultSystemAddress  string = "localhost:9090"
	defaultPostgresParams        = "host=localhost user=postgres password=postgres dbname=test_db sslmode=disable"
)

// Flags аргументы программы
var Flags flags

// ParseArgs функция для чтения аргументов программы
func ParseArgs() {
	flag.StringVar(&Flags.AccrualAddress, "a", defaultServerAddress, "accrual address")
	flag.StringVar(&Flags.ServerAddress, "r", defaultSystemAddress, "server address")
	flag.StringVar(&Flags.DBConf, "d", defaultPostgresParams, "db params")
	flag.Parse()

	serverAddressEnv, findAddress := os.LookupEnv("RUN_ADDRESS")
	accrualAddress, findSystemAddress := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	DBConf, findDBConf := os.LookupEnv("DATABASE_URI")

	if findAddress {
		Flags.AccrualAddress = accrualAddress
	}
	if findSystemAddress {
		Flags.ServerAddress = serverAddressEnv
	}
	if findDBConf {
		Flags.DBConf = DBConf
	}
	logger.Log.Info(
		fmt.Sprintf("Parse argument's is done, accrual addres: [%s], server addres: [%s], db url: [%s]",
			Flags.AccrualAddress, Flags.ServerAddress, Flags.DBConf),
	)
}

// HasFlagOrEnvPostgresVariable проверка наличия env переменной
func HasFlagOrEnvPostgresVariable() bool {
	_, find := os.LookupEnv("DATABASE_URI")
	if Flags.DBConf != defaultPostgresParams || find {
		return true
	}
	return false
}
