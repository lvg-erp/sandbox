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

// Создать сеанс без мест
func (r *sessionRepo) CreateFilmSession(filmID, cinemaID int, start time.Time) (int, error) {
	var sessionID int
	err := r.db.QueryRow(`
        INSERT INTO film_sessions (film_id, cinema_id, start_time)
        VALUES ($1, $2, $3) RETURNING id`, filmID, cinemaID, start).Scan(&sessionID)
	return sessionID, err
}

func (r *sessionRepo) GetAllSessions() ([]entities.SessionInfo, error) {
	rows, err := r.db.Query(`
        SELECT fs.id, f.title, fs.start_time
        FROM film_sessions fs
        JOIN films f ON fs.film_id = f.id
        ORDER BY fs.start_time`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []entities.SessionInfo
	for rows.Next() {
		var s entities.SessionInfo
		if err := rows.Scan(&s.ID, &s.FilmTitle, &s.StartTime); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *sessionRepo) ListFilmsForCinema(cinemaID int) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
        SELECT fs.id, f.title as film_title, f.duration, fs.start_time, c.name as cinema_name 
        FROM film_sessions fs 
        JOIN films f ON fs.film_id = f.id 
        JOIN cinemas c ON fs.cinema_id = c.id 
        WHERE fs.cinema_id = $1`, cinemaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	for rows.Next() {
		var s struct {
			ID         int
			FilmTitle  string
			Duration   int
			StartTime  string
			CinemaName string
		}
		err := rows.Scan(&s.ID, &s.FilmTitle, &s.Duration, &s.StartTime, &s.CinemaName)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, map[string]interface{}{
			"id":          s.ID,
			"film_title":  s.FilmTitle,
			"duration":    s.Duration,
			"start_time":  s.StartTime,
			"cinema_name": s.CinemaName,
		})
	}
	return sessions, nil
}

func (r *sessionRepo) ListCinemasWithSessions() ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
        SELECT c.id, c.name, COUNT(fs.id) as session_count
        FROM cinemas c LEFT JOIN film_sessions fs ON c.id = fs.cinema_id
        GROUP BY c.id, c.name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cinemas []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var count int
		err := rows.Scan(&id, &name, &count)
		if err != nil {
			return nil, err
		}
		cinemas = append(cinemas, map[string]interface{}{
			"id":            id,
			"name":          name,
			"session_count": float64(count), // для JSON float64
		})
	}
	return cinemas, nil
}
