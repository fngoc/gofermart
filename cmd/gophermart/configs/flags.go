package configs

import (
	"flag"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"os"
)

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

var Flags flags

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

func HasFlagOrEnvPostgresVariable() bool {
	_, find := os.LookupEnv("DATABASE_URI")
	if Flags.DBConf != defaultPostgresParams || find {
		return true
	}
	return false
}
