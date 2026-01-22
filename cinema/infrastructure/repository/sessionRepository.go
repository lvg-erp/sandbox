package repository

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"database/sql"
	"encoding/hex"
	"errors"
	"golang.org/x/exp/rand"
	"time"
)

type sessionRepo struct {
	db *sql.DB
}

// NewRepo — конструктор
func NewSessionRepo(db *sql.DB) ports.SessionRepository {
	return &sessionRepo{db: db}
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession — создаёт сессию в БД

func (r *sessionRepo) CreateSession(userID int, expires time.Time) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}
	_, err = r.db.Exec(`
        INSERT INTO sessions (user_id, token, expires_at)
        VALUES ($1, $2, $3)`,
		userID, token, expires,
	)
	return token, err
}

func (r *sessionRepo) GetSession(token string) (entities.Session, error) {
	var s entities.Session
	err := r.db.QueryRow(`
        SELECT id, user_id, token, expires_at FROM sessions WHERE token = $1`, token,
	).Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt)
	if err != nil {
		return s, err
	}
	if time.Now().After(s.ExpiresAt) {
		r.DeleteSession(token)
		return s, errors.New("session expired")
	}
	return s, nil
}

func (r *sessionRepo) DeleteSession(token string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE token = $1`, token)
	return err
}
