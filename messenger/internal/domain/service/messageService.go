package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"
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
	_, err := s.chatRepo.GetByUUID(ctx, chatUUID)
	if err != nil {
		return nil, errors.New("chat not found")
	}

	// Проверяем, что отправитель участник чата
	participants, err := s.chatRepo.GetParticipants(ctx, chatUUID)
	if err != nil {
		return nil, err
	}

	isParticipant := false
	for _, p := range participants {
		if p.UUID == senderUUID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return nil, errors.New("sender is not a participant of this chat")
	}

	message := &entity.Message{
		UUID:       uuid.New(),
		ChatUUID:   chatUUID,
		SenderUUID: senderUUID,
		Body:       body,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *MessageService) GetChatMessages(ctx context.Context, chatUUID uuid.UUID, limit, offset int) ([]*entity.Message, error) {
	return s.messageRepo.GetChatMessages(ctx, chatUUID, limit, offset)
}

func (s *MessageService) DeleteMessage(ctx context.Context, messageUUID uuid.UUID) error {
	return s.messageRepo.Delete(ctx, messageUUID)
}
