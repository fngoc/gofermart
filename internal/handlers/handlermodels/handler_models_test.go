package handlermodels

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthRequestJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        AuthRequest
		expectedJSON string
	}{
		{
			name: "Valid AuthRequest",
			input: AuthRequest{
				Login:    "userName",
				Password: "password",
			},
			expectedJSON: `{"login":"userName","password":"password"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Проверка маршалинга (структура -> JSON)
			marshaled, err := json.Marshal(tt.input)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expectedJSON, string(marshaled))

			// Проверка анмаршалинга (JSON -> структура)
			var unmarshaled AuthRequest
			err = json.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, unmarshaled)
		})
	}
}

func TestWithdrawRequestJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        WithdrawRequest
		expectedJSON string
	}{
		{
			name: "Valid WithdrawRequest",
			input: WithdrawRequest{
				Order: "12345",
				Sum:   100.50,
			},
			expectedJSON: `{"order":"12345","sum":100.50}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Проверка маршалинга (структура -> JSON)
			marshaled, err := json.Marshal(tt.input)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expectedJSON, string(marshaled))

			// Проверка анмаршалинга (JSON -> структура)
			var unmarshaled WithdrawRequest
			err = json.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, unmarshaled)
		})
	}
}
