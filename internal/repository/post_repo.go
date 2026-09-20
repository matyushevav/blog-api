package repository

import (
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PostRepo struct {
	db *sql.DB
}

func NewPostRepo(db *sql.DB) *PostRepo {
	return &PostRepo{db: db}
}

// Create создает новый пост в БД.
func (r *PostRepo) Create(ctx context.Context, post *model.Post) error {
	const query = `
		INSERT INTO public.posts (title, content, author_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	post.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		post.Title,
		post.Content,
		post.AuthorID,
		post.CreatedAt,
	).Scan(&post.ID)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	return nil
}

// GetByID получает пост по id из БД.
func (r *PostRepo) GetByID(ctx context.Context, id int) (*model.Post, error) {
	const query = `
		SELECT id, title, content, author_id, created_at
		FROM public.posts
		WHERE id = $1`

	post := &model.Post{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}

	return post, nil
}

// GetAll получает посты из БД с сортировкой по дате создания.
func (r *PostRepo) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	const query = `
		SELECT id, title, content, author_id, created_at
		FROM public.posts
		ORDER BY created_at DESC
		LIMIT $1
		OFFSET $2`

	posts := make([]*model.Post, 0, limit)

	res, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer res.Close()
	for res.Next() {
		post := &model.Post{}

		err := res.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		posts = append(posts, post)
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate posts: %w", err)
	}

	return posts, nil
}

// GetTotalCount получает общее количество постов.
func (r *PostRepo) GetTotalCount(ctx context.Context) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM public.posts`

	var count int

	err := r.db.QueryRowContext(ctx, query).Scan(
		&count,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to get total count of posts: %w", err)
	}

	return count, nil
}

// Exists проверяет существование поста по id.
func (r *PostRepo) Exists(ctx context.Context, id int) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM public.posts WHERE id = $1)`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&exists,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check post by id: %w", err)
	}

	return exists, nil
}

// GetByAuthorID получает посты по id автора.
func (r *PostRepo) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	const query = `
		SELECT id, title, content, author_id, created_at
		FROM public.posts
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		OFFSET $3`

	posts := make([]*model.Post, 0, limit)

	res, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by authorID: %w", err)
	}
	defer res.Close()
	for res.Next() {
		post := &model.Post{}

		err := res.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		posts = append(posts, post)
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate posts: %w", err)
	}

	return posts, nil
}

// GetTotalCountByAuthorID получает общее количество постов автора.
func (r *PostRepo) GetTotalCountByAuthorID(ctx context.Context, authorID int) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM public.posts
		WHERE author_id = $1`

	var count int

	err := r.db.QueryRowContext(ctx, query, authorID).Scan(
		&count,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to get total count of posts by authorID: %w", err)
	}

	return count, nil
}
