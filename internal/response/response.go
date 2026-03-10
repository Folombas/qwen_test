// Package response предоставляет утилиты для формирования HTTP ответов
package response

import (
	"encoding/json"
	"net/http"
)

// JSON отправляет JSON ответ с указанным статус кодом
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}

// Error отправляет JSON ответ с ошибкой
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

// Success отправляет успешный JSON ответ
func Success(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"message": message})
}

// Data отправляет JSON ответ с данными
func Data(w http.ResponseWriter, status int, data interface{}) {
	JSON(w, status, data)
}

// MethodNotAllowed отправляет ошибку метода
func MethodNotAllowed(w http.ResponseWriter, method string) {
	Error(w, http.StatusMethodNotAllowed, "Method "+method+" not allowed")
}

// BadRequest отправляет ошибку запроса
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// Unauthorized отправляет ошибку авторизации
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message)
}

// Forbidden отправляет ошибку доступа
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message)
}

// NotFound отправляет ошибку не найдено
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message)
}

// InternalServerError отправляет ошибку сервера
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message)
}
