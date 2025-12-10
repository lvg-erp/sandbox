package usecases

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
)

type AuthUseCase struct {
	SessionRepo ports.SessionRepository
	UserRepo    ports.UserRepository
}

func (uc *AuthUseCase) ExecuteAuth(token string) (*entities.User, error) {
	session, err := uc.SessionRepo.GetSession(token)
	if err != nil {
		return nil, err
	}
	user, err := uc.UserRepo.GetUserByID(session.UserID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
