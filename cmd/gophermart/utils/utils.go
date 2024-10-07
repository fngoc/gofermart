package utils

import (
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"time"
	"unicode"
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

// CheckLunAlg проверяет корректность номера по алгоритму Луна
func CheckLunAlg(number string) bool {
	var sum int
	double := false

	// Идем с конца строки к началу
	for i := len(number) - 1; i >= 0; i-- {
		r := rune(number[i])

		if !unicode.IsDigit(r) {
			continue
		}

		digit := int(r - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}
	return sum%10 == 0
}
