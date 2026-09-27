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

// commentListResponse - ответ со страницей комментариев и общим количеством.
type commentListResponse struct {
	Comments []model.CommentResponse `json:"comments"`
	Total    int                     `json:"total"`
}

type CommentHandler struct {
	commentService service.CommentServiceInterface
	eventLogger    *logger.EventLogger
}

// NewCommentHandler создает новый экземпляр CommentHandler
func NewCommentHandler(commentService service.CommentServiceInterface, eventLogger *logger.EventLogger) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		eventLogger:    eventLogger,
	}
}

// Create обрабатывает POST /api/posts/{postId}/comments
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		h.respondWithError(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	postID, err := strconv.Atoi(chi.URLParam(r, "postId"))
	if err != nil || postID <= 0 {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)

		return
	}

	var req model.CommentCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid JSON body", http.StatusBadRequest)

		return
	}

	req.PostID = postID

	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)

		return
	}

	comment, err := h.commentService.Create(r.Context(), userID, postID, req.Content)
	if err != nil {
		HandleServiceError(w, err)

		return
	}

	h.eventLogger.LogEvent(fmt.Sprintf("user %d created comment %d", userID, comment.ID))

	resp := model.CommentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		PostID:    comment.PostID,
		AuthorID:  comment.AuthorID,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	h.respondWithJSON(w, resp, http.StatusCreated)
}

// GetByPost обрабатывает GET /api/posts/{postId}/comments?limit=10&offset=0
// TODO: Извлечь postID, получить limit и offset, валидировать,
// вызвать commentService.GetByPost(). Вернуть список комментариев и общее количество
func (h *CommentHandler) GetByPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	postID, err := strconv.Atoi(chi.URLParam(r, "postId"))
	if err != nil || postID <= 0 {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)

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

	comments, total, err := h.commentService.GetByPost(r.Context(), postID, limit, offset)
	if err != nil {
		HandleServiceError(w, err)

		return
	}

	items := make([]model.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		items = append(items, model.CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			PostID:    comment.PostID,
			AuthorID:  comment.AuthorID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		})
	}

	h.respondWithJSON(w, commentListResponse{Comments: items, Total: total}, http.StatusOK)
}

// respondWithJSON отправляет JSON ответ
// TODO: Установить Content-Type, WriteHeader, закодировать JSON
func (h *CommentHandler) respondWithJSON(w http.ResponseWriter, data interface{}, statusCode int) {
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
func (h *CommentHandler) respondWithError(w http.ResponseWriter, message string, statusCode int) {
	h.respondWithJSON(w, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}, statusCode)
}
