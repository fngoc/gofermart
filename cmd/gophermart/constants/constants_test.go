package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderStatusConstants(t *testing.T) {
	t.Run("Test New order status", func(t *testing.T) {
		expectedStatus := orderStatus("NEW")
		assert.Equal(t, expectedStatus, New, "Expected 'New' status to be 'NEW'")
	})
}

func TestContextKeyConstants(t *testing.T) {
	t.Run("Test UserNameKey context key", func(t *testing.T) {
		expectedKey := contextKey("userName")
		assert.Equal(t, expectedKey, UserNameKey, "Expected 'UserNameKey' to be 'userName'")
	})
}
