package repository

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"database/sql"
	"errors"
)

// === ФИЛЬМЫ ===

type filmRepo struct {
	db *sql.DB
}

// NewRepo — конструктор
func NewFilmRepo(db *sql.DB) ports.FilmRepository {
	return &filmRepo{db: db}
}

// CreateFilm — добавить фильм
func (r *filmRepo) CreateFilm(f entities.Film) error {
	_, err := r.db.Exec(`
		INSERT INTO films (title, poster, description, duration)
		VALUES ($1, $2, $3, $4)`,
		f.Title, f.Poster, f.Description, f.Duration,
	)
	return err
}

// ListFilms — все фильмы
func (r *filmRepo) ListFilms() ([]entities.Film, error) {
	rows, err := r.db.Query(`
		SELECT id, title, poster, description, duration
		FROM films ORDER BY added_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []entities.Film
	for rows.Next() {
		var f entities.Film
		if err := rows.Scan(&f.ID, &f.Title, &f.Poster, &f.Description, &f.Duration); err != nil {
			return nil, err
		}
		films = append(films, f)
	}
	return films, rows.Err()
}

// GetFilm — по ID
func (r *filmRepo) GetFilm(id int) (entities.Film, error) {
	var f entities.Film
	err := r.db.QueryRow(`
		SELECT id, title, poster, description, duration
		FROM films WHERE id = $1`, id,
	).Scan(&f.ID, &f.Title, &f.Poster, &f.Description, &f.Duration)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return f, errors.New("film not found")
		}
		return f, err
	}
	return f, nil
}
