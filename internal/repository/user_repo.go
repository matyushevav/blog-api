package repository

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type UserRepo struct {
	db *sql.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create сохраняет нового пользователя и проставляет ему ID, выданный БД.
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	const query = `
		INSERT INTO public.users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID получает данные пользователя по id из БД.
func (r *UserRepo) GetByID(ctx context.Context, id int) (*model.User, error) {
	const query = `
		SELECT id, username, email, password, created_at, updated_at
		FROM public.users
		WHERE id = $1`

	user := &model.User{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail получает данные пользователя по email из БД.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const query = `
		SELECT id, username, email, password, created_at, updated_at
		FROM public.users
		WHERE email = $1`

	user := &model.User{}

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetByUsername получает данные пользователя по username из БД.
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	const query = `
		SELECT id, username, email, password, created_at, updated_at
		FROM public.users
		WHERE username = $1`

	user := &model.User{}

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// ExistsByEmail проверяет существование пользователя в БД по email.
func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM public.users WHERE email = $1)`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&exists,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check user by email: %w", err)
	}

	return exists, nil
}

// ExistsByUsername проверяет существование пользователя в БД по username.
func (r *UserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&exists,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check user by username: %w", err)
	}

	return exists, nil
}

// Update обновляет данные о пользователе в БД.
func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	const query = `
		UPDATE public.users
		SET username = $1, email = $2, password = $3, updated_at = $4
		WHERE id = $5`

	user.UpdatedAt = time.Now()

	res, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user with id=%d: %w", user.ID, apperrors.ErrUserNotFound)
	}

	return nil
}

// Delete удаляет пользователя из БД.
func (r *UserRepo) Delete(ctx context.Context, id int) error {
	const query = `
		DELETE FROM public.users
		WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user with id=%d: %w", id, apperrors.ErrUserNotFound)
	}

	return nil
}
