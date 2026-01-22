package repository

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"database/sql"
	"time"
)

type sessionFilmRepo struct {
	db *sql.DB
}

func NewSessionFilmRepo(db *sql.DB) ports.SessionFilmRepository {
	return &sessionFilmRepo{db: db}
}

// Создать сеанс без мест
func (r *sessionFilmRepo) CreateFilmSession(filmID, cinemaID int, start, end time.Time) (int, error) {
	var sessionID int
	err := r.db.QueryRow(`
        INSERT INTO film_sessions (
            film_id, 
            cinema_id, 
            date, 
            start_time, 
            end_time, 
            start_time_full
        )
        VALUES ($1, $2, $3::date, $3::time, $4::time, $3)
        RETURNING id`,
		filmID, cinemaID, start, end,
	).Scan(&sessionID)
	return sessionID, err
}

func (r *sessionFilmRepo) GetFilmSessions(filmID, cinemaID int) ([]entities.SessionFilm, error) {
	rows, err := r.db.Query(`
        SELECT 
            fs.id,
            f.title AS film_title,
            f.duration,
            fs.start_time,
            fs.end_time
        FROM film_sessions fs 
        JOIN films f ON fs.film_id = f.id 
        JOIN cinemas c ON fs.cinema_id = c.id 
        WHERE fs.film_id = $1 
          AND fs.cinema_id = $2`,
		filmID, cinemaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []entities.SessionFilm

	for rows.Next() {
		var s entities.SessionFilm

		err := rows.Scan(
			&s.ID,
			&s.FilmTitle,
			&s.Duration, // ← было пропущено
			&s.StartTime,
			&s.EndTime,
		)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

func (r *sessionFilmRepo) GetAllFilmsSessions() ([]entities.SessionFilm, error) {
	rows, err := r.db.Query(`
        SELECT fs.id, f.title, fs.start_time
        FROM film_sessions fs
        JOIN films f ON fs.film_id = f.id
        ORDER BY fs.start_time`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []entities.SessionFilm
	for rows.Next() {
		var s entities.SessionFilm
		if err := rows.Scan(&s.ID, &s.FilmTitle, &s.StartTime, &s.EndTime); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *sessionFilmRepo) ListFilmsForCinema(cinemaID int) ([]map[string]interface{}, error) {
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

func (r *sessionFilmRepo) ListCinemasWithSessions() ([]map[string]interface{}, error) {
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
