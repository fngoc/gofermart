package middlewares

import (
	"bytes"
	"compress/gzip"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGzipMiddleware тестирует работу GzipMiddleware для входящих и исходящих данных.
func TestGzipMiddleware(t *testing.T) {
	// тестовый HTTP-обработчик, который мы будем оборачивать в GzipMiddleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Hello, World!"))
		assert.NoError(t, err)
	})

	// Тест 1: сжатие запроса и ответа
	reqBody := compressRequestBody(t, "Test Gzip Request Body")
	req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()

	// Оборачиваем тестовый обработчик через middleware
	handler := GzipMiddleware(testHandler)
	handler.ServeHTTP(rr, req)

	// Декомпрессируем ответ и проверяем содержимое
	decompressedBody := decompressResponseBody(t, rr.Body)
	assert.Equal(t, "Hello, World!", decompressedBody)

	// Тест 2: запрос без сжатия, но с поддержкой сжатия ответа
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Декомпрессируем ответ и проверяем содержимое
	decompressedBody = decompressResponseBody(t, rr.Body)
	assert.Equal(t, "Hello, World!", decompressedBody)

	// Тест 3: запрос без сжатия и без поддержки сжатия ответа
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Проверяем, что заголовок Content-Encoding отсутствует
	assert.Equal(t, "", rr.Header().Get("Content-Encoding"))

	// Проверяем содержимое ответа без декомпрессии
	assert.Equal(t, "Hello, World!", rr.Body.String())
}

// compressRequestBody сжимает тело запроса с использованием gzip
func compressRequestBody(t *testing.T, body string) io.Reader {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(body))
	assert.NoError(t, err)
	err = zw.Close()
	assert.NoError(t, err)
	return &buf
}

// decompressResponseBody декомпрессирует ответ, сжатый с использованием gzip
func decompressResponseBody(t *testing.T, body *bytes.Buffer) string {
	zr, err := gzip.NewReader(body)
	assert.NoError(t, err)
	defer zr.Close()

	var decompressedBody bytes.Buffer
	_, err = io.Copy(&decompressedBody, zr)
	assert.NoError(t, err)

	return decompressedBody.String()
}
