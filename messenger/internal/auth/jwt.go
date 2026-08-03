package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type JWTConfig struct {
	SecretKey     string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type Claims struct {
	UserUUID string `json:"user_uuid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewJWTConfig(secretKey string) *JWTConfig {
	return &JWTConfig{
		SecretKey:     secretKey,
		AccessExpiry:  24 * time.Hour,     // Access токен живет 24 часа
		RefreshExpiry: 7 * 24 * time.Hour, // Refresh токен живет 7 дней
	}
}

// GenerateAccessToken генерирует access токен
func (c *JWTConfig) GenerateAccessToken(userUUID uuid.UUID, username string) (string, error) {
	claims := Claims{
		UserUUID: userUUID.String(),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "messenger",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.SecretKey))
}

// GenerateRefreshToken генерирует refresh токен
func (c *JWTConfig) GenerateRefreshToken() (string, error) {
	// Используем UUID как refresh токен
	return uuid.New().String(), nil
}

// ValidateToken проверяет токен и возвращает claims
func (c *JWTConfig) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.SecretKey), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Проверяем срок действия
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// GetUserUUIDFromToken извлекает UUID пользователя из токена
func (c *JWTConfig) GetUserUUIDFromToken(tokenString string) (uuid.UUID, error) {
	claims, err := c.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	userUUID, err := uuid.Parse(claims.UserUUID)
	if err != nil {
		return uuid.Nil, errors.New("invalid user UUID in token")
	}

	return userUUID, nil
}
