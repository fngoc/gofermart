package configs

import (
	"flag"
	"fmt"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"os"
)

type flags struct {
	ServerAddress string
	SystemAddress string
	DBConf        string
}

const (
	defaultServerAddress  string = "localhost:8080"
	defaultSystemAddress  string = "localhost:9090"
	defaultPostgresParams        = "host=localhost user=postgres password=postgres dbname=test_db sslmode=disable"
)

var Flags flags

func ParseArgs() {
	flag.StringVar(&Flags.ServerAddress, "a", defaultServerAddress, "server address")
	flag.StringVar(&Flags.SystemAddress, "r", defaultSystemAddress, "system address")
	flag.StringVar(&Flags.DBConf, "d", defaultPostgresParams, "db params")
	flag.Parse()

	serverAddressEnv, findAddress := os.LookupEnv("RUN_ADDRESS")
	systemAddress, findSystemAddress := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	DBConf, findDBConf := os.LookupEnv("DATABASE_URI")

	if findAddress {
		Flags.ServerAddress = serverAddressEnv
	}
	if findSystemAddress {
		Flags.SystemAddress = systemAddress
	}
	if findDBConf {
		Flags.DBConf = DBConf
	}
	logger.Log.Info(
		fmt.Sprintf("Parse argument's is done, server addres: [%s], system addres: [%s], db url: [%s]",
			Flags.ServerAddress, Flags.SystemAddress, Flags.DBConf),
	)
}

func HasFlagOrEnvPostgresVariable() bool {
	_, find := os.LookupEnv("DATABASE_URI")
	if Flags.DBConf != defaultPostgresParams || find {
		return true
	}
	return false
}
