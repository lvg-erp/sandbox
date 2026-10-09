package service

import (
	"context"
	"errors"
	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"
	"time"

	"github.com/google/uuid"
)

type MessageService struct {
	messageRepo repository.MessageRepository
	chatRepo    repository.ChatRepository
	userRepo    repository.UserRepository
}

func NewMessageService(messageRepo repository.MessageRepository, chatRepo repository.ChatRepository, userRepo repository.UserRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
		userRepo:    userRepo,
	}
}

func (s *MessageService) SendMessage(ctx context.Context, chatUUID, senderUUID uuid.UUID, body string) (*entity.Message, error) {
	if body == "" {
		return nil, errors.New("message body cannot be empty")
	}

	// Проверяем существование чата
	chat, err := s.chatRepo.GetByUUID(ctx, chatUUID)
	if err != nil || chat == nil {
		return nil, errors.New("chat not found")
	}

	// Проверяем, что отправитель участник чата
	isParticipant, err := s.chatRepo.IsParticipant(ctx, chatUUID, senderUUID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, errors.New("sender is not a participant of this chat")
	}

	now := time.Now() // ВАЖНО: сохраняем текущее время

	message := &entity.Message{
		UUID:       uuid.New(),
		ChatUUID:   chatUUID,
		SenderUUID: senderUUID,
		Body:       body,
		CreatedAt:  now, // Устанавливаем время
		UpdatedAt:  now,
		Deleted:    false,
		Edited:     false,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *MessageService) GetChatMessages(ctx context.Context, chatUUID, userUUID uuid.UUID, limit, offset int) ([]*entity.Message, error) {
	// Читать историю может только участник чата
	isParticipant, err := s.chatRepo.IsParticipant(ctx, chatUUID, userUUID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, errors.New("access denied")
	}

	return s.messageRepo.GetChatMessages(ctx, chatUUID, limit, offset)
}

func (s *MessageService) DeleteMessage(ctx context.Context, messageUUID uuid.UUID) error {
	return s.messageRepo.Delete(ctx, messageUUID)
}
