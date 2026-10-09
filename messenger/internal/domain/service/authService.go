package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"messanger/internal/auth"
	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"

	"github.com/google/uuid"
)

var ErrInvalidRefreshToken = errors.New("invalid refresh token")

type AuthService struct {
	userRepo      repository.UserRepository
	tokenRepo     repository.TokenRepository
	refreshExpiry time.Duration
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, refreshExpiry time.Duration) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		refreshExpiry: refreshExpiry,
	}
}

// hashToken возвращает SHA-256 токена: в БД хранится хэш,
// поэтому утечка таблицы не позволяет угнать сессии.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// IssueRefreshToken генерирует случайный токен и сохраняет его хэш в БД
// с ограниченным сроком жизни.
func (s *AuthService) IssueRefreshToken(ctx context.Context, userUUID uuid.UUID) (string, error) {
	raw, err := auth.GenerateRandomToken()
	if err != nil {
		return "", err
	}

	now := time.Now()
	rt := &entity.RefreshToken{
		UUID:      uuid.New(),
		Token:     hashToken(raw),
		UserUUID:  userUUID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.refreshExpiry),
	}

	if err := s.tokenRepo.Create(ctx, rt); err != nil {
		return "", err
	}

	return raw, nil
}

// RefreshTokens проверяет refresh-токен, отзывает его и выдает новый (ротация).
// Повторное использование уже отозванного токена считается компрометацией
// и отзывает все токены пользователя.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*entity.User, string, error) {
	if refreshToken == "" {
		return nil, "", ErrInvalidRefreshToken
	}

	hashed := hashToken(refreshToken)

	rt, err := s.tokenRepo.GetByToken(ctx, hashed)
	if err != nil {
		return nil, "", err
	}
	if rt == nil {
		return nil, "", ErrInvalidRefreshToken
	}

	if rt.Revoked {
		_ = s.tokenRepo.RevokeAllForUser(ctx, rt.UserUUID)
		return nil, "", ErrInvalidRefreshToken
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, "", ErrInvalidRefreshToken
	}

	user, err := s.userRepo.GetByUUID(ctx, rt.UserUUID)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrInvalidRefreshToken
	}

	if err := s.tokenRepo.Revoke(ctx, hashed); err != nil {
		return nil, "", err
	}

	newToken, err := s.IssueRefreshToken(ctx, user.UUID)
	if err != nil {
		return nil, "", err
	}

	return user, newToken, nil
}

func (s *AuthService) RevokeToken(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.Revoke(ctx, hashToken(refreshToken))
}
