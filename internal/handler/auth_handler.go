package handler

import (
	"advanced-blog-management-system/internal/middleware"
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
func NewAuthHandler(userService service.UserServiceInterface) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register обрабатывает POST /api/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	var req model.UserCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid JSON body", http.StatusBadRequest)

		return
	}

	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)

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
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	var req model.UserLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid JSON body", http.StatusBadRequest)

		return
	}

	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)

		return
	}

	resp, err := h.userService.Login(r.Context(), &req)
	if err != nil {
		HandleServiceError(w, err)

		return
	}

	h.respondWithJSON(w, resp, http.StatusOK)
}

// GetProfile получает профиль текущего пользователя
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		h.respondWithError(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		HandleServiceError(w, err)

		return
	}

	h.respondWithJSON(w, user.ToResponse(), http.StatusOK)
}

// respondWithJSON отправляет JSON ответ с заданным статус кодом
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
func (h *AuthHandler) respondWithError(w http.ResponseWriter, message string, statusCode int) {
	h.respondWithJSON(w, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}, statusCode)
}
