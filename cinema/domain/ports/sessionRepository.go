package ports

import (
	"cinema/domain/entities"
	"time"
)

type SessionRepository interface {
	GetAllSessions() ([]entities.SessionInfo, error)
	CreateSession(userID int, expires time.Time) (string, error)
	GetSession(token string) (entities.Session, error)
	DeleteSession(token string) error
	CreateFilmSession(filmID, cinemaID int, start time.Time) (int, error)
	ListFilmsForCinema(cinemaID int) ([]map[string]interface{}, error)
	ListCinemasWithSessions() ([]map[string]interface{}, error)
}
