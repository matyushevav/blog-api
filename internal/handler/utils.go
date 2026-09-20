package handler

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// ErrorResponse представляет структуру ошибки в ответе
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// TODO: Реализовать WriteError(w http.ResponseWriter, message string, statusCode int)
// 1. Установить Content-Type: application/json
// 2. Установить статус код через WriteHeader()
// 3. Закодировать ErrorResponse в JSON используя json.NewEncoder()
// ErrorResponse должен содержать Error = http.StatusText(statusCode) и Message = message
//
// WriteError отправляет ошибку в формате JSON с заданным статус кодом.
func WriteError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// TODO: Реализовать HandleServiceError(w http.ResponseWriter, err error)
// 1. Проверить тип ошибки используя errors.Is()
// 2. Для каждого типа ошибки из apperrors вернуть нужный HTTP статус:
//    - ErrUserAlreadyExists -> 409 Conflict
//    - ErrInvalidCredentials -> 401 Unauthorized
//    - ErrPostNotFound -> 404 Not Found
//    - ErrCommentNotFound -> 404 Not Found
//    - ErrForbidden -> 403 Forbidden
//    - ErrUnauthorized -> 401 Unauthorized
// 3. Проверить validator.ValidationErrors -> 400 Bad Request
// 4. Для других ошибок -> 500 Internal Server Error
// 5. Использовать WriteError() для отправки ответа с нужным сообщением
//
// HandleServiceError переводит ошибку сервиса в подходящий HTTP статус.
func HandleServiceError(w http.ResponseWriter, err error) {
	var validationErrs validator.ValidationErrors

	switch {
	case errors.Is(err, apperrors.ErrUserAlreadyExists):
		WriteError(w, apperrors.ErrUserAlreadyExists.Error(), http.StatusConflict)
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		WriteError(w, apperrors.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
	case errors.Is(err, apperrors.ErrUnauthorized):
		WriteError(w, apperrors.ErrUnauthorized.Error(), http.StatusUnauthorized)
	case errors.Is(err, apperrors.ErrForbidden):
		WriteError(w, apperrors.ErrForbidden.Error(), http.StatusForbidden)
	case errors.Is(err, apperrors.ErrUserNotFound):
		WriteError(w, apperrors.ErrUserNotFound.Error(), http.StatusNotFound)
	case errors.Is(err, apperrors.ErrPostNotFound):
		WriteError(w, apperrors.ErrPostNotFound.Error(), http.StatusNotFound)
	case errors.Is(err, apperrors.ErrCommentNotFound):
		WriteError(w, apperrors.ErrCommentNotFound.Error(), http.StatusNotFound)
	case errors.Is(err, apperrors.ErrInvalidPostID):
		WriteError(w, apperrors.ErrInvalidPostID.Error(), http.StatusBadRequest)
	case errors.As(err, &validationErrs):
		WriteError(w, validationErrs.Error(), http.StatusBadRequest)
	default:
		log.Printf("Unhandled service error: %v", err)
		WriteError(w, "internal server error", http.StatusInternalServerError)
	}
}
