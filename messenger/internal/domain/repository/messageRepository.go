package repository

import (
	"context"
	"github.com/google/uuid"
	"messanger/internal/domain/entity"
)

type MessageRepository interface {
	Create(ctx context.Context, message *entity.Message) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*entity.Message, error)
	GetChatMessages(ctx context.Context, chatUUID uuid.UUID, limit, offset int) ([]*entity.Message, error)
	Delete(ctx context.Context, uuid uuid.UUID) error
	GetLastMessages(ctx context.Context, chatUUID uuid.UUID, limit int) ([]*entity.Message, error)
}
