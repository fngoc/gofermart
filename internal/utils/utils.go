package utils

import (
	"time"

	"github.com/fngoc/gofermart/internal/logger"
)

// ConvertTime конвертор времени в нужный формат
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
