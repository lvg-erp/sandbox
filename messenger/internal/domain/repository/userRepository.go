package repository

import (
	"context"
	"github.com/google/uuid"
	"messanger/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdateLastSeen(ctx context.Context, uuid uuid.UUID) error
	GetAll(ctx context.Context) ([]*entity.User, error)
}
