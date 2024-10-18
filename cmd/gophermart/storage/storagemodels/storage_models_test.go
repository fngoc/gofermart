package storagemodels

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        Order
		expectedJSON string
	}{
		{
			name: "Valid Order with Accrual",
			input: Order{
				Number:     "123456",
				Status:     "PROCESSED",
				Accrual:    150.75,
				UploadedAt: "2024-10-18T12:00:00Z",
			},
			expectedJSON: `{"number":"123456","status":"PROCESSED","accrual":150.75,"uploaded_at":"2024-10-18T12:00:00Z"}`,
		},
		{
			name: "Order without Accrual",
			input: Order{
				Number:     "654321",
				Status:     "NEW",
				UploadedAt: "2024-10-18T12:30:00Z",
			},
			expectedJSON: `{"number":"654321","status":"NEW","uploaded_at":"2024-10-18T12:30:00Z"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Проверка маршалинга (структура -> JSON)
			marshaled, err := json.Marshal(tt.input)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expectedJSON, string(marshaled))

			// Проверка анмаршалинга (JSON -> структура)
			var unmarshaled Order
			err = json.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, unmarshaled)
		})
	}
}

func TestBalanceJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        Balance
		expectedJSON string
	}{
		{
			name: "Valid Balance",
			input: Balance{
				Current:   500.75,
				Withdrawn: 100.25,
			},
			expectedJSON: `{"current":500.75,"withdrawn":100.25}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Проверка маршалинга (структура -> JSON)
			marshaled, err := json.Marshal(tt.input)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expectedJSON, string(marshaled))

			// Проверка анмаршалинга (JSON -> структура)
			var unmarshaled Balance
			err = json.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, unmarshaled)
		})
	}
}

func TestTransactionJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        Transaction
		expectedJSON string
	}{
		{
			name: "Valid Transaction",
			input: Transaction{
				OrderNumber: "987654",
				Sum:         200.50,
				ProcessedAt: "2024-10-18T14:00:00Z",
			},
			expectedJSON: `{"order":"987654","sum":200.50,"processed_at":"2024-10-18T14:00:00Z"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Проверка маршалинга (структура -> JSON)
			marshaled, err := json.Marshal(tt.input)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expectedJSON, string(marshaled))

			// Проверка анмаршалинга (JSON -> структура)
			var unmarshaled Transaction
			err = json.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, unmarshaled)
		})
	}
}
