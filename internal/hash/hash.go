package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashingPassword функция для хэширования строки
func HashingPassword(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:]), nil
}
