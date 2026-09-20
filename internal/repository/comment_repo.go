package repository

import (
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type CommentRepo struct {
	db *sql.DB
}

func NewCommentRepo(db *sql.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

// Create создает новый комментарий в БД.
func (r *CommentRepo) Create(ctx context.Context, comment *model.Comment) error {
	const query = `
		INSERT INTO public.comments (content, post_id, author_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		comment.Content,
		comment.PostID,
		comment.AuthorID,
		comment.CreatedAt,
		comment.UpdatedAt,
	).Scan(&comment.ID)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// GetByID получает комментарий по id из БД.
func (r *CommentRepo) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	const query = `
		SELECT id, content, post_id, author_id, created_at, updated_at
		FROM public.comments
		WHERE id = $1`

	comment := &model.Comment{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.Content,
		&comment.PostID,
		&comment.AuthorID,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}

	return comment, nil
}

// GetByPostID получает комментарии по postID.
func (r *CommentRepo) GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, error) {
	const query = `
		SELECT id, content, post_id, author_id, created_at, updated_at
		FROM public.comments
		WHERE post_id = $1
		ORDER BY created_at ASC
		LIMIT $2
		OFFSET $3`

	comments := make([]*model.Comment, 0, limit)

	res, err := r.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by postID: %w", err)
	}
	defer res.Close()
	for res.Next() {
		comment := &model.Comment{}

		err := res.Scan(
			&comment.ID,
			&comment.Content,
			&comment.PostID,
			&comment.AuthorID,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comments: %w", err)
		}

		comments = append(comments, comment)
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate comments: %w", err)
	}

	return comments, nil
}

// GetCountByPostID получает общее количество комментариев поста
func (r *CommentRepo) GetCountByPostID(ctx context.Context, postID int) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM public.comments
		WHERE post_id = $1`

	var count int

	err := r.db.QueryRowContext(ctx, query, postID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count of comments by postID: %w", err)
	}

	return count, nil
}
