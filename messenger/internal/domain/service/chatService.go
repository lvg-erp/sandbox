package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"
)

type ChatService struct {
	chatRepo    repository.ChatRepository
	userRepo    repository.UserRepository
	messageRepo repository.MessageRepository
}

func NewChatService(chatRepo repository.ChatRepository, userRepo repository.UserRepository, messageRepo repository.MessageRepository) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		userRepo:    userRepo,
		messageRepo: messageRepo,
	}
}

func (s *ChatService) CreatePersonalChat(ctx context.Context, user1UUID, user2UUID uuid.UUID) (*entity.Chat, error) {
	// Проверяем существование пользователей
	_, err := s.userRepo.GetByUUID(ctx, user1UUID)
	if err != nil {
		return nil, errors.New("user1 not found")
	}

	_, err = s.userRepo.GetByUUID(ctx, user2UUID)
	if err != nil {
		return nil, errors.New("user2 not found")
	}

	// Проверяем существующий чат
	existingChat, err := s.chatRepo.GetPersonalChat(ctx, user1UUID, user2UUID)
	if err == nil && existingChat != nil {
		return existingChat, nil
	}

	// Создаем новый чат
	return s.chatRepo.CreatePersonalChat(ctx, user1UUID, user2UUID)
}

func (s *ChatService) GetUserChats(ctx context.Context, userUUID uuid.UUID) ([]*entity.Chat, error) {
	return s.chatRepo.GetUserChats(ctx, userUUID)
}

func (s *ChatService) GetChatWithLastMessage(ctx context.Context, chatUUID uuid.UUID, userUUID uuid.UUID) (map[string]interface{}, error) {
	chat, err := s.chatRepo.GetByUUID(ctx, chatUUID)
	if err != nil {
		return nil, err
	}

	if chat == nil {
		return nil, errors.New("chat not found")
	}

	messages, err := s.messageRepo.GetLastMessages(ctx, chatUUID, 1)
	if err != nil {
		return nil, err
	}

	participants, err := s.chatRepo.GetParticipants(ctx, chatUUID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"chat":         chat,
		"participants": participants,
	}

	if len(messages) > 0 {
		result["last_message"] = messages[0]
	}

	return result, nil
}
