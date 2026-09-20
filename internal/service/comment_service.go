package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"context"
	"fmt"
	"strings"
)

type CommentService struct {
	repo     *repository.CommentRepo
	postRepo repository.PostRepository
}

func NewCommentService(repo *repository.CommentRepo, postRepo repository.PostRepository) *CommentService {
	return &CommentService{
		repo:     repo,
		postRepo: postRepo,
	}
}

// Create создает новый комментарий.
func (s *CommentService) Create(ctx context.Context, userID, postID int, content string) (*model.Comment, error) {
	trimContent := strings.TrimSpace(content)

	// Правила длины и postID > 0 описаны тегами в CommentCreateRequest
	req := &model.CommentCreateRequest{Content: trimContent, PostID: postID}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid create comment request: %w", err)
	}

	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	if !exists {
		return nil, apperrors.ErrPostNotFound
	}

	comment := &model.Comment{Content: trimContent, PostID: postID, AuthorID: userID}
	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

// GetByPost получает комментарии к посту.
func (s *CommentService) GetByPost(ctx context.Context, postID, limit, offset int) ([]*model.Comment, int, error) {
	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get post: %w", err)
	}

	if !exists {
		return nil, 0, apperrors.ErrPostNotFound
	}

	limit, offset = normalizePagination(limit, offset)

	comments, err := s.repo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get comments: %w", err)
	}

	count, err := s.repo.GetCountByPostID(ctx, postID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count of comments: %w", err)
	}

	return comments, count, nil
}
