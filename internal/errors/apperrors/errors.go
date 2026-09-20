package apperrors

import "errors"

var (
	// ErrUnauthorized - пользователь не аутентифицирован (нет или невалиден токен).
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden - пользователь аутентифицирован, но не имеет прав на действие.
	ErrForbidden = errors.New("forbidden")

	// ErrUserNotFound - пользователь не найден в БД.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists - пользователь с таким email или username уже существует.
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidCredentials - неверный email или пароль при входе.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrPostNotFound - пост не найден в БД.
	ErrPostNotFound = errors.New("post not found")

	// ErrCommentNotFound - комментарий не найден в БД.
	ErrCommentNotFound = errors.New("comment not found")

	// ErrInvalidPostID - передан невалидный ID поста.
	ErrInvalidPostID = errors.New("invalid post id")
)

