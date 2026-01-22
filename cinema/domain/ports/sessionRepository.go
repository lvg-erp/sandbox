package ports

import (
	"cinema/domain/entities"
	"time"
)

type SessionRepository interface {
	CreateSession(userID int, expires time.Time) (string, error)
	GetSession(token string) (entities.Session, error)
	DeleteSession(token string) error
}
