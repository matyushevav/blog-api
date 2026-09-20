package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"advanced-blog-management-system/pkg/auth"
	"context"
	"fmt"
)

type UserService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewUserService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register регистрирует нового пользователя и сразу выдает ему JWT токен.
func (s *UserService) Register(ctx context.Context, req *model.UserCreateRequest) (*model.TokenResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid register request: %w", err)
	}

	emailTaken, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	if emailTaken {
		return nil, apperrors.ErrUserAlreadyExists
	}

	usernameTaken, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}

	if usernameTaken {
		return nil, apperrors.ErrUserAlreadyExists
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hash,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToResponse(),
	}, nil
}

// Login осуществляет авторизацию пользователя в системе.
func (s *UserService) Login(ctx context.Context, req *model.UserLoginRequest) (*model.TokenResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid login request: %w", err)
	}

	userByEmail, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if userByEmail == nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if !auth.CheckPassword(req.Password, userByEmail.Password) {
		return nil, apperrors.ErrInvalidCredentials
	}

	token, expiresAt, err := s.jwtManager.GenerateToken(
		userByEmail.ID,
		userByEmail.Email,
		userByEmail.Username,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      userByEmail.ToResponse(),
	}, nil
}

// GetByID возвращает пользователя по ID.
func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}

	return user, nil
}

// GetByEmail возвращает пользователя по email.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}

	return user, nil
}
