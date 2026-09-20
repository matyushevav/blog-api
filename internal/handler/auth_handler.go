package handler

import (
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"
	"encoding/json"
	"log"
	"net/http"
)

type AuthHandler struct {
	userService service.UserServiceInterface
}

// NewAuthHandler создает новый экземпляр AuthHandler
// TODO: Инициализировать с userService (интерфейс)
func NewAuthHandler(userService service.UserServiceInterface) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register обрабатывает POST /api/register
// TODO: Проверить метод, распарсить JSON, валидировать, вызвать userService.Register()
// Вернуть TokenResponse со статусом 201 или ошибку с нужным кодом
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	var req model.UserCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "invalid JSON body", http.StatusBadRequest)

		return
	}

	if err := req.Validate(); err != nil {
		WriteError(w, err.Error(), http.StatusBadRequest)

		return
	}

	resp, err := h.userService.Register(r.Context(), &req)
	if err != nil {
		HandleServiceError(w, err)

		return
	}

	h.respondWithJSON(w, resp, http.StatusCreated)
}

// Login обрабатывает POST /api/login
// TODO: Проверить метод, распарсить JSON, валидировать, вызвать userService.Login()
// Вернуть TokenResponse со статусом 200 или ошибку с нужным кодом
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// TODO: реализовать
}

// GetProfile получает профиль текущего пользователя
// TODO: Проверить метод GET, получить userID из контекста, вернуть UserResponse (опционально)
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// TODO: реализовать получение профиля (опционально)
}

// respondWithJSON отправляет JSON ответ с заданным статус кодом
// TODO: Установить Content-Type, WriteHeader, закодировать JSON
func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// respondWithError отправляет JSON ошибку
// TODO: Создать ErrorResponse и отправить используя respondWithJSON()
func (h *AuthHandler) respondWithError(w http.ResponseWriter, message string, statusCode int) {
	h.respondWithJSON(w, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}, statusCode)
}
