package usecases

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"time"
)

type CreateUserSessionInput struct {
	UserID  int
	Expires time.Time
}

type CreateUserSessionOutput struct {
	Token string
}

type GetUserSessionInput struct {
	Token string
}

type GetUserSessionOutput struct {
	Session entities.Session
}

type DeleteUserSessionInput struct {
	Token string
}

type CreateFilmSessionInput struct {
	FilmID   int
	CinemaID int
	Start    time.Time
}

type CreateFilmSessionOutput struct {
	ID int
}

type ListFilmSessionsForCinemaInput struct {
	CinemaID int
}

type ListFilmSessionsForCinemaOutput []map[string]interface{}

type ListCinemasWithSessionsOutput []map[string]interface{}

type SessionUseCase struct {
	Repo ports.SessionRepository // Интерфейс для сессий (user/film)
}

func (uc *SessionUseCase) ExecuteCreateUserSession(input *CreateUserSessionInput) (*CreateUserSessionOutput, error) {
	token, err := uc.Repo.CreateSession(input.UserID, input.Expires)
	if err != nil {
		return nil, err
	}
	return &CreateUserSessionOutput{Token: token}, nil
}

func (uc *SessionUseCase) ExecuteGetUserSession(input *GetUserSessionInput) (*GetUserSessionOutput, error) {
	session, err := uc.Repo.GetSession(input.Token)
	if err != nil {
		return nil, err
	}
	return &GetUserSessionOutput{Session: session}, nil
}

func (uc *SessionUseCase) ExecuteDeleteUserSession(input *DeleteUserSessionInput) error {
	return uc.Repo.DeleteSession(input.Token)
}

func (uc *SessionUseCase) ExecuteCreateFilmSession(input *CreateFilmSessionInput) (*CreateFilmSessionOutput, error) {
	id, err := uc.Repo.CreateFilmSession(input.FilmID, input.CinemaID, input.Start)
	if err != nil {
		return nil, err
	}
	return &CreateFilmSessionOutput{ID: id}, nil
}

func (uc *SessionUseCase) ExecuteGetAllSessions() ([]entities.SessionInfo, error) {
	return uc.Repo.GetAllSessions()
}

func (uc *SessionUseCase) ExecuteListSessionsForCinema(input *ListFilmSessionsForCinemaInput) (ListFilmSessionsForCinemaOutput, error) {
	sessions, err := uc.Repo.ListFilmsForCinema(input.CinemaID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (uc *SessionUseCase) ExecuteListCinemasWithSessions() (ListCinemasWithSessionsOutput, error) {
	cinemas, err := uc.Repo.ListCinemasWithSessions()
	if err != nil {
		return nil, err
	}
	return cinemas, nil
}
