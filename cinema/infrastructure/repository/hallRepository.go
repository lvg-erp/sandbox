package repository

import (
	"cinema/domain/ports"
	"database/sql"
)

type hallRepo struct {
	db *sql.DB
}

// NewRepo — конструктор
func NewHallRepo(db *sql.DB) ports.HallRepository {
	return &hallRepo{db: db}
}

func (r *hallRepo) CreateHall(cinemaID uint, name string, rows, seatsPerRow int) (uint, error) {
	var hallID uint
	err := r.db.QueryRow(`
		INSERT INTO halls (cinema_id, name, rows, seats_per_row) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`,
		cinemaID, name, rows, seatsPerRow,
	).Scan(&hallID)
	return hallID, err
}
