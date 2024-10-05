package utils

import (
	"database/sql"
	"github.com/fngoc/gofermart/cmd/gophermart/logger"
	"time"
	"unicode"
)

func GetValueByNullInt64(value sql.NullInt64) int {
	if value.Valid {
		return int(value.Int64)
	}
	return 0
}

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
	// Маркер для удвоения каждой второй цифры
	double := false

	// Идем с конца строки к началу
	for i := len(number) - 1; i >= 0; i-- {
		// Получаем текущий символ
		r := rune(number[i])

		// Пропускаем нечисловые символы
		if !unicode.IsDigit(r) {
			continue
		}

		// Преобразуем символ в целое число
		digit := int(r - '0')

		// Если нужно удвоить каждую вторую цифру
		if double {
			digit *= 2
			// Если результат больше 9, вычитаем 9
			if digit > 9 {
				digit -= 9
			}
		}

		// Добавляем к общей сумме
		sum += digit

		// Меняем флаг удвоения для следующей цифры
		double = !double
	}

	// Проверяем, делится ли сумма на 10
	return sum%10 == 0
}
