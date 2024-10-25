package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertTime(t *testing.T) {
	tests := []struct {
		name         string
		inputTime    string
		expectedTime string
	}{
		{
			name:         "Valid RFC3339Nano time conversion",
			inputTime:    "2023-10-10T15:04:05.000000000Z",
			expectedTime: "2023-10-10T18:04:05+03:00",
		},
		{
			name:         "Valid RFC3339Nano with fractional seconds",
			inputTime:    "2023-10-10T15:04:05.123456789Z",
			expectedTime: "2023-10-10T18:04:05+03:00",
		},
		{
			name:         "Invalid time format (return original string)",
			inputTime:    "invalid-time-format",
			expectedTime: "invalid-time-format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertedTime := ConvertTime(tt.inputTime)
			assert.Equal(t, tt.expectedTime, convertedTime)
		})
	}
}
