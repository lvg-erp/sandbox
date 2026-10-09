package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"messanger/internal/domain/entity"
)

type fakeTokenRepo struct {
	tokens map[string]*entity.RefreshToken // ключ — хэш токена
}

func newFakeTokenRepo() *fakeTokenRepo {
	return &fakeTokenRepo{tokens: make(map[string]*entity.RefreshToken)}
}

func (f *fakeTokenRepo) Create(_ context.Context, t *entity.RefreshToken) error {
	f.tokens[t.Token] = t
	return nil
}

func (f *fakeTokenRepo) GetByToken(_ context.Context, token string) (*entity.RefreshToken, error) {
	if t, ok := f.tokens[token]; ok {
		return t, nil
	}
	return nil, nil
}

func (f *fakeTokenRepo) Revoke(_ context.Context, token string) error {
	if t, ok := f.tokens[token]; ok {
		t.Revoked = true
	}
	return nil
}

func (f *fakeTokenRepo) RevokeAllForUser(_ context.Context, userUUID uuid.UUID) error {
	for _, t := range f.tokens {
		if t.UserUUID == userUUID {
			t.Revoked = true
		}
	}
	return nil
}

func newAuthServiceWithFakes() (*AuthService, *fakeTokenRepo, uuid.UUID) {
	users := newFakeUserRepo()
	userID := uuid.New()
	users.users["alice"] = &entity.User{UUID: userID, Username: "alice"}
	tokens := newFakeTokenRepo()
	return NewAuthService(users, tokens, time.Hour), tokens, userID
}

func TestRefreshTokensRotatesAndRevokesOld(t *testing.T) {
	svc, tokens, userID := newAuthServiceWithFakes()
	ctx := context.Background()

	old, err := svc.IssueRefreshToken(ctx, userID)
	if err != nil {
		t.Fatalf("IssueRefreshToken: %v", err)
	}

	user, newToken, err := svc.RefreshTokens(ctx, old)
	if err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}
	if user.Username != "alice" {
		t.Fatalf("unexpected user: %s", user.Username)
	}
	if newToken == old {
		t.Fatal("token was not rotated")
	}
	if !tokens.tokens[hashToken(old)].Revoked {
		t.Fatal("old token was not revoked")
	}

	// Старый токен больше не работает
	if _, _, err := svc.RefreshTokens(ctx, old); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("want ErrInvalidRefreshToken for rotated token, got %v", err)
	}
}

func TestRefreshTokenReuseRevokesAllUserTokens(t *testing.T) {
	svc, tokens, userID := newAuthServiceWithFakes()
	ctx := context.Background()

	old, err := svc.IssueRefreshToken(ctx, userID)
	if err != nil {
		t.Fatalf("IssueRefreshToken: %v", err)
	}
	_, newToken, err := svc.RefreshTokens(ctx, old)
	if err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}

	// Попытка повторно использовать отозванный токен
	if _, _, err := svc.RefreshTokens(ctx, old); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("want ErrInvalidRefreshToken, got %v", err)
	}
	// Все токены пользователя отозваны
	if !tokens.tokens[hashToken(newToken)].Revoked {
		t.Fatal("reuse detection must revoke all user tokens")
	}
}

func TestRefreshUnknownToken(t *testing.T) {
	svc, _, _ := newAuthServiceWithFakes()

	if _, _, err := svc.RefreshTokens(context.Background(), "unknown-token"); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("want ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshExpiredToken(t *testing.T) {
	svc, tokens, userID := newAuthServiceWithFakes()

	expired := &entity.RefreshToken{
		Token:     hashToken("expired"),
		UserUUID:  userID,
		ExpiresAt: time.Now().Add(-time.Minute),
	}
	tokens.tokens[expired.Token] = expired

	if _, _, err := svc.RefreshTokens(context.Background(), "expired"); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("want ErrInvalidRefreshToken for expired token, got %v", err)
	}
}

func TestRevokeToken(t *testing.T) {
	svc, tokens, userID := newAuthServiceWithFakes()
	ctx := context.Background()

	raw, err := svc.IssueRefreshToken(ctx, userID)
	if err != nil {
		t.Fatalf("IssueRefreshToken: %v", err)
	}

	if err := svc.RevokeToken(ctx, raw); err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}
	if !tokens.tokens[hashToken(raw)].Revoked {
		t.Fatal("token was not revoked")
	}
}
