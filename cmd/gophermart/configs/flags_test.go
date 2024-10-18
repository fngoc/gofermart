package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasFlagOrEnvPostgresVariable(t *testing.T) {
	tests := []struct {
		name           string
		envVariable    string
		flagDBConf     string
		expectedResult bool
	}{
		{
			name:           "Default flag with no env variable",
			envVariable:    "",
			flagDBConf:     defaultPostgresParams,
			expectedResult: false,
		},
		{
			name:           "Env variable set",
			envVariable:    "host=localhost user=custom password=custom dbname=custom_db sslmode=disable",
			flagDBConf:     defaultPostgresParams,
			expectedResult: true,
		},
		{
			name:           "Flag value set and no env variable",
			envVariable:    "",
			flagDBConf:     "host=localhost user=test password=test dbname=test_db sslmode=disable",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сбрасываем флаг перед каждым тестом
			Flags.DBConf = tt.flagDBConf

			// Устанавливаем переменную окружения
			if tt.envVariable != "" {
				os.Setenv("DATABASE_URI", tt.envVariable)
			} else {
				os.Unsetenv("DATABASE_URI")
			}

			// Проверяем результат
			result := HasFlagOrEnvPostgresVariable()
			assert.Equal(t, tt.expectedResult, result)

			// Чистим переменную окружения после теста
			os.Unsetenv("DATABASE_URI")
		})
	}
}
