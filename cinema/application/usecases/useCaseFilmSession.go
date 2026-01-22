package usecases

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"time"
)

type GetFilmSessionsInput struct {
	FilmID   int `json:"film_id"`
	CinemaID int `json:"cinema_id"`
}

type CreateFilmSessionInput struct {
	FilmID   int
	CinemaID int
	Start    time.Time
	End      time.Time
}

type CreateFilmSessionOutput struct {
	ID int
}

type ListFilmSessionsForCinemaInput struct {
	CinemaID int
}

type ListFilmSessionsForCinemaOutput []map[string]interface{}

type SessionFilmUseCase struct {
	Repo ports.SessionFilmRepository // Интерфейс для сессий (user/film)
}

func (uc *SessionFilmUseCase) ExecuteGetFilmSessions(input *GetFilmSessionsInput) ([]entities.SessionFilm, error) {
	fs, err := uc.Repo.GetFilmSessions(input.FilmID, input.CinemaID)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func (uc *SessionFilmUseCase) ExecuteCreateFilmSession(input *CreateFilmSessionInput) (*CreateFilmSessionOutput, error) {
	id, err := uc.Repo.CreateFilmSession(input.FilmID, input.CinemaID, input.Start, input.End)
	if err != nil {
		return nil, err
	}
	return &CreateFilmSessionOutput{ID: id}, nil
}

func (uc *SessionFilmUseCase) ExecuteGetAllSessions() ([]entities.SessionFilm, error) {
	return uc.Repo.GetAllFilmsSessions()
}

func (uc *SessionFilmUseCase) ExecuteListSessionsForCinema(input *ListFilmSessionsForCinemaInput) (ListFilmSessionsForCinemaOutput, error) {
	sessions, err := uc.Repo.ListFilmsForCinema(input.CinemaID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (uc *SessionFilmUseCase) ExecuteListCinemasWithSessions() (ListCinemasWithSessionsOutput, error) {
	cinemas, err := uc.Repo.ListCinemasWithSessions()
	if err != nil {
		return nil, err
	}
	return cinemas, nil
}
