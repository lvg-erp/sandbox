package repository

import (
	"context"

	"github.com/google/uuid"

	"messanger/internal/domain/entity"
)

type TokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllForUser(ctx context.Context, userUUID uuid.UUID) error
}
