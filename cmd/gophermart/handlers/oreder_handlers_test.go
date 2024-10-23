package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadOrderWebhook_Success(t *testing.T) {
	mockStore := &mockStorage{
		GetUserNameByOrderIDFunc: func(orderID int) string {
			return ""
		},
		GetUserIDByNameFunc: func(userName string) (int, error) {
			return 1, nil
		},
		CreateOrderFunc: func(userID int, orderID int) error {
			return nil
		},
	}

	storage.SetDbInstance(mockStore)

	orderID := 79927398713
	requestBody, _ := json.Marshal(orderID)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), constants.UserNameKey, "test_user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	LoadOrderWebhook(w, req)

	if status := w.Code; status != http.StatusAccepted {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}
}

func TestLoadOrderWebhook_Conflict(t *testing.T) {
	mockStore := &mockStorage{
		GetUserNameByOrderIDFunc: func(orderID int) string {
			return "another_user"
		},
	}

	storage.SetDbInstance(mockStore)

	orderID := 79927398713
	requestBody, _ := json.Marshal(orderID)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), constants.UserNameKey, "test_user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	LoadOrderWebhook(w, req)

	if status := w.Code; status != http.StatusConflict {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusConflict)
	}
}

func TestLoadOrderWebhook_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	w := httptest.NewRecorder()

	LoadOrderWebhook(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestLoadOrderWebhook_InvalidContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoadOrderWebhook(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestLoadOrderWebhook_LuhnCheckFail(t *testing.T) {
	orderID := 123 // Некорректный orderID для алгоритма Луна
	requestBody, _ := json.Marshal(orderID)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), constants.UserNameKey, "test_user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	LoadOrderWebhook(w, req)

	if status := w.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}
}
