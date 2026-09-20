package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"context"
	"fmt"
)

const (
	// defaultLimit - размер страницы, если limit некорректный
	defaultLimit = 10

	// maxLimit - верхняя граница размера страницы
	maxLimit = 100
)

type PostService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// Create создает пост.
func (s *PostService) Create(ctx context.Context, userID int, req *model.PostCreateRequest) (*model.Post, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid create post request: %w", err)
	}

	// Проверяем автора
	author, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post author: %w", err)
	}

	if author == nil {
		return nil, apperrors.ErrUserNotFound
	}

	post := &model.Post{Title: req.Title, Content: req.Content, AuthorID: userID}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

// GetByID получает пост по ID.
func (s *PostService) GetByID(ctx context.Context, id int, requestorID int) (*model.Post, error) {
	// requestorID не используется: все посты публичные
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	if post == nil {
		return nil, apperrors.ErrPostNotFound
	}
	return post, nil

}

// GetAll получает слайс всех постов с количеством.
func (s *PostService) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, int, error) {
	limit, offset = normalizePagination(limit, offset)

	posts, err := s.postRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all posts: %w", err)
	}

	count, err := s.postRepo.GetTotalCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count of posts: %w", err)
	}

	return posts, count, nil
}

// GetByAuthor получает слайс всех постов пользователя с количеством.
func (s *PostService) GetByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, int, error) {
	limit, offset = normalizePagination(limit, offset)

	posts, err := s.postRepo.GetByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user posts: %w", err)
	}

	count, err := s.postRepo.GetTotalCountByAuthorID(ctx, authorID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count of user posts: %w", err)
	}

	return posts, count, nil
}

// normalizePagination приводит limit и offset к допустимым значениям.
func normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultLimit
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	if offset < 0 {
		offset = 0
	}

	return limit, offset
}
