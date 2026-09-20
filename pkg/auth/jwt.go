package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

const (
	// tokenIssuer - значение поля iss, кто выпустил токен.
	tokenIssuer = "blog-api"

	// defaultTTLHours - срок жизни токена, если в конфиге задан некорректный.
	defaultTTLHours = 24
)

// Claims представляет данные, хранимые в JWT токене.
type Claims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTManager хранит секретный ключ для подписи токенов и срок их жизни.
type JWTManager struct {
	secretKey []byte
	ttl       time.Duration
}

// NewJWTManager создает менеджер токенов с заданным секретом и сроком жизни.
func NewJWTManager(secretKey string, ttlHours int) *JWTManager {
	if ttlHours <= 0 {
		ttlHours = defaultTTLHours
	}

	return &JWTManager{
		secretKey: []byte(secretKey),
		ttl:       time.Duration(ttlHours) * time.Hour,
	}
}

// GenerateToken выпускает подписанный JWT и возвращает его вместе со временем истечения.
func (m *JWTManager) GenerateToken(userID int, email, username string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)

	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    tokenIssuer,
			Subject:   strconv.Itoa(userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, expiresAt, nil
}

// ValidateToken проверяет подпись и срок действия токена и возвращает его Claims.
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return m.secretKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}

		return nil, ErrInvalidToken
	}

	if token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt == nil || !claims.ExpiresAt.After(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}
