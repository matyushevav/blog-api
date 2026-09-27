package handler

import (
	"advanced-blog-management-system/internal/logger"
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// postListResponse - ответ со страницей постов и общим количеством.
type postListResponse struct {
	Posts []model.PostResponse `json:"posts"`
	Total int                  `json:"total"`
}

type PostHandler struct {
	postService service.PostServiceInterface
	eventLogger *logger.EventLogger
}

// NewPostHandler создает новый экземпляр PostHandler
func NewPostHandler(postService service.PostServiceInterface, eventLogger *logger.EventLogger) *PostHandler {
	return &PostHandler{
		postService: postService,
		eventLogger: eventLogger,
	}
}

// Create обрабатывает POST /api/posts
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		h.respondWithError(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	var req model.PostCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid JSON body", http.StatusBadRequest)

		return
	}

	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)

		return
	}

	post, err := h.postService.Create(r.Context(), userID, &req)
	if err != nil {
		HandleServiceError(w, err)

		return
	}

	h.eventLogger.LogEvent(fmt.Sprintf("user %d created post %d", userID, post.ID))

	resp := model.PostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		AuthorID:  post.AuthorID,
		CreatedAt: post.CreatedAt,
	}

	h.respondWithJSON(w, resp, http.StatusCreated)
}

// GetByID обрабатывает GET /api/posts/{id}
func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	postID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || postID <= 0 {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)

		return
	}

	// Эндпоинт публичный: без токена userID будет 0
	userID, _ := middleware.GetUserIDFromContext(r.Context())

	post, err := h.postService.GetByID(r.Context(), postID, userID)
	if err != nil {
		HandleServiceError(w, err)

		return
	}

	resp := model.PostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		AuthorID:  post.AuthorID,
		CreatedAt: post.CreatedAt,
	}

	h.respondWithJSON(w, resp, http.StatusOK)
}

// GetAll обрабатывает GET /api/posts?limit=10&offset=0
func (h *PostHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	limit, err := queryInt(r, "limit")
	if err != nil {
		h.respondWithError(w, "invalid limit", http.StatusBadRequest)

		return
	}

	offset, err := queryInt(r, "offset")
	if err != nil {
		h.respondWithError(w, "invalid offset", http.StatusBadRequest)

		return
	}

	posts, total, err := h.postService.GetAll(r.Context(), limit, offset)
	if err != nil {
		HandleServiceError(w, err)

		return
	}

	items := make([]model.PostResponse, 0, len(posts))
	for _, post := range posts {
		items = append(items, model.PostResponse{
			ID:        post.ID,
			Title:     post.Title,
			Content:   post.Content,
			AuthorID:  post.AuthorID,
			CreatedAt: post.CreatedAt,
		})
	}

	h.respondWithJSON(w, postListResponse{Posts: items, Total: total}, http.StatusOK)
}

// GetByAuthor обрабатывает GET /api/posts/author/{authorID}?limit=10&offset=0
func (h *PostHandler) GetByAuthor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	authorID, err := strconv.Atoi(chi.URLParam(r, "authorID"))
	if err != nil || authorID <= 0 {
		h.respondWithError(w, "invalid author id", http.StatusBadRequest)

		return
	}

	limit, err := queryInt(r, "limit")
	if err != nil {
		h.respondWithError(w, "invalid limit", http.StatusBadRequest)

		return
	}

	offset, err := queryInt(r, "offset")
	if err != nil {
		h.respondWithError(w, "invalid offset", http.StatusBadRequest)

		return
	}

	posts, total, err := h.postService.GetByAuthor(r.Context(), authorID, limit, offset)
	if err != nil {
		HandleServiceError(w, err)

		return
	}

	items := make([]model.PostResponse, 0, len(posts))
	for _, post := range posts {
		items = append(items, model.PostResponse{
			ID:        post.ID,
			Title:     post.Title,
			Content:   post.Content,
			AuthorID:  post.AuthorID,
			CreatedAt: post.CreatedAt,
		})
	}

	h.respondWithJSON(w, postListResponse{Posts: items, Total: total}, http.StatusOK)
}

// respondWithJSON отправляет JSON ответ
func (h *PostHandler) respondWithJSON(w http.ResponseWriter, data interface{}, statusCode int) {
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
func (h *PostHandler) respondWithError(w http.ResponseWriter, message string, statusCode int) {
	h.respondWithJSON(w, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}, statusCode)
}
