package hash

import (
	"testing"
)

func TestHashingPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected string
		hasError bool
	}{
		{
			name:     "Valid password",
			password: "password123",
			expected: "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f",
			hasError: false,
		},
		{
			name:     "Empty password",
			password: "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			hasError: false,
		},
		{
			name:     "Special characters password",
			password: "!@#$%^&*()",
			expected: "95ce789c5c9d18490972709838ca3a9719094bca3ac16332cfec0652b0236141",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashingPassword(tt.password)
			if (err != nil) != tt.hasError {
				t.Errorf("HashingPassword() error = %s, expected error = %v", err, tt.hasError)
			}
			if hash != tt.expected {
				t.Errorf("HashingPassword() = %s, expected = %v", hash, tt.expected)
			}
		})
	}
}
