package utils

import (
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"time"
)

func ConvertTime(t string) string {
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
