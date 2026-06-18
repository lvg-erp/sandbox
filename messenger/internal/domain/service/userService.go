package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, username string) (*entity.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	user := &entity.User{
		UUID:     uuid.New(),
		Username: username,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, uuid uuid.UUID) (*entity.User, error) {
	return s.userRepo.GetByUUID(ctx, uuid)
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *UserService) UpdateLastSeen(ctx context.Context, uuid uuid.UUID) error {
	return s.userRepo.UpdateLastSeen(ctx, uuid)
}
