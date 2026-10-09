package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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
		AccessExpiry:  15 * time.Minute,   // короткий access, обновляется refresh-токеном
		RefreshExpiry: 7 * 24 * time.Hour, // refresh хранится в БД и отзывается
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

// GenerateRandomToken возвращает криптостойкий случайный токен (32 байта, hex).
// Используется для refresh-токенов, которые хранятся в БД и отзываются.
func GenerateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidateToken проверяет токен и возвращает claims
func (c *JWTConfig) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Разрешаем только HMAC, чтобы нельзя было подменить алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
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
