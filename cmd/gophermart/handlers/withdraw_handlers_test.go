package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/fngoc/gofermart/cmd/gophermart/constants"
	"github.com/fngoc/gofermart/cmd/gophermart/storage"
	"github.com/fngoc/gofermart/cmd/gophermart/storage/storagemodels"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListWithdrawalsBalanceWebhook_Success(t *testing.T) {
	mockTransactions := []storagemodels.Transaction{
		{OrderNumber: "12345", Sum: 100.50, ProcessedAt: "2024-10-22"},
	}

	// Мокаем хранилище
	mockStore := &mockStorage{
		GetUserIDByNameFunc: func(userName string) (int, error) {
			return 1, nil
		},
		GetAllTransactionByUserIDFunc: func(userID int) ([]storagemodels.Transaction, error) {
			return mockTransactions, nil
		},
	}

	// Подменяем хранилище в пакете storage
	storage.SetDbInstance(mockStore)

	// Создаем запрос и ответ
	req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
	// Устанавливаем контекст с валидным токеном пользователя
	ctx := context.WithValue(req.Context(), constants.UserNameKey, "test_user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Вызываем тестируемую функцию
	ListWithdrawalsBalanceWebhook(w, req)

	// Проверяем код ответа
	if status := w.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем тело ответа
	var response []storagemodels.Transaction
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if len(response) != len(mockTransactions) {
		t.Errorf("Handler returned unexpected number of transactions: got %v want %v", len(response), len(mockTransactions))
	}
}

func TestListWithdrawalsBalanceWebhook_NoTransactions(t *testing.T) {
	mockStore := &mockStorage{
		GetUserIDByNameFunc: func(userName string) (int, error) {
			return 1, nil
		},
		GetAllTransactionByUserIDFunc: func(userID int) ([]storagemodels.Transaction, error) {
			return []storagemodels.Transaction{}, nil
		},
	}

	storage.SetDbInstance(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
	ctx := context.WithValue(req.Context(), constants.UserNameKey, "test_user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	ListWithdrawalsBalanceWebhook(w, req)

	if status := w.Code; status != http.StatusNoContent {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestListWithdrawalsBalanceWebhook_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/withdrawals", nil)
	w := httptest.NewRecorder()

	ListWithdrawalsBalanceWebhook(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestListWithdrawalsBalanceWebhook_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
	w := httptest.NewRecorder()

	// Контекст без валидного токена
	req = req.WithContext(context.Background())

	ListWithdrawalsBalanceWebhook(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestListWithdrawalsBalanceWebhook_InternalServerError(t *testing.T) {
	// Мокаем хранилище с ошибкой
	mockStore := &mockStorage{
		GetUserIDByNameFunc: func(userName string) (int, error) {
			return 0, errors.New("database error")
		},
	}

	storage.SetDbInstance(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
	ctx := context.WithValue(req.Context(), constants.UserNameKey, "test_user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	ListWithdrawalsBalanceWebhook(w, req)

	if status := w.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}
