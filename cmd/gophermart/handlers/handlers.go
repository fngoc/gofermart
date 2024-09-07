package handlers

import (
	"net/http"
)

// TestWebhook тестовый обработчик GET HTTP-запроса
func TestWebhook(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
