package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
)

const MinPasswordLength = 8

var ErrInvalidCredentials = errors.New("invalid username or password")

// dummyHash нужен, чтобы Authenticate тратил одинаковое время на ответ,
// когда пользователь не найден и когда пароль неверный.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("messenger-timing-equalizer"), bcrypt.DefaultCost)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, username, password string) (*entity.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if len(password) < MinPasswordLength {
		return nil, fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &entity.User{
		UUID:         uuid.New(),
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
		LastSeen:     now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate проверяет имя пользователя и пароль.
// И "пользователь не найден", и "пароль неверный" возвращают ErrInvalidCredentials,
// чтобы по ошибке нельзя было понять, существует ли логин.
func (s *UserService) Authenticate(ctx context.Context, username, password string) (*entity.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
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
