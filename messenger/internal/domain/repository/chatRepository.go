package repository

import (
	"context"
	"github.com/google/uuid"
	"messanger/internal/domain/entity"
)

type ChatRepository interface {
	Create(ctx context.Context, chat *entity.Chat) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*entity.Chat, error)
	GetUserChats(ctx context.Context, userUUID uuid.UUID) ([]*entity.Chat, error)
	AddParticipant(ctx context.Context, chatUUID, userUUID uuid.UUID) error
	GetParticipants(ctx context.Context, chatUUID uuid.UUID) ([]*entity.User, error)
	GetPersonalChat(ctx context.Context, user1UUID, user2UUID uuid.UUID) (*entity.Chat, error)
	CreatePersonalChat(ctx context.Context, user1UUID, user2UUID uuid.UUID) (*entity.Chat, error)
}
