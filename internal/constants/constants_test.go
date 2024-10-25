package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
	}{
		{
			name:     "Test New order status",
			expected: "NEW",
			actual:   New,
		},
		{
			name:     "Test Processed order status",
			expected: "PROCESSED",
			actual:   Processed,
		},
		{
			name:     "Test Invalid order status",
			expected: "INVALID",
			actual:   Invalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.actual, "Expected status to match")
		})
	}
}

func TestContextKeyConstants(t *testing.T) {
	t.Run("Test UserNameKey context key", func(t *testing.T) {
		expectedKey := contextKey("userName")
		assert.Equal(t, expectedKey, UserNameKey, "Expected 'UserNameKey' to match")
	})
}
