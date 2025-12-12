package repository

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"database/sql"
	"errors"
	"log"
)

type cinemaRepo struct {
	db *sql.DB
}

// NewRepo — конструктор
func NewCinemaRepo(db *sql.DB) ports.CinemaRepository {
	return &cinemaRepo{db: db}
}

// CreateCinema — добавить кинотеатр
func (r *cinemaRepo) CreateCinema(cinema entities.Cinema) error {
	_, err := r.db.Exec(`
		INSERT INTO cinemas (name, address, city, phone, total_seats, poster)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		cinema.Name, cinema.Address, cinema.City, cinema.Phone, cinema.TotalSeats, cinema.Poster,
	)
	return err
}

// ListCinemas — все кинотеатры
func (r *cinemaRepo) ListCinemas() ([]entities.Cinema, error) {
	rows, err := r.db.Query(`
       SELECT id, name, address, city, phone, total_seats, poster, created_at
       FROM cinemas ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Error closing rows: %v", closeErr)
		}
	}()

	var cinemas []entities.Cinema
	for rows.Next() {
		var c entities.Cinema
		if err := rows.Scan(&c.ID, &c.Name, &c.Address, &c.City, &c.Phone, &c.TotalSeats, &c.Poster, &c.CreatedAt); err != nil {
			return nil, err
		}
		cinemas = append(cinemas, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cinemas, nil
}

// GetCinema — по ID
func (r *cinemaRepo) GetCinema(id int) (entities.Cinema, error) {
	var c entities.Cinema
	err := r.db.QueryRow(`
		SELECT id, name, address, city, phone, total_seats, poster, created_at
		FROM cinemas WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Address, &c.City, &c.Phone, &c.TotalSeats, &c.Poster, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c, errors.New("cinema not found")
		}
		return c, err
	}
	return c, nil
}
