package repository

import (
	"cinema/domain/entities"
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

func (r *hallRepo) ListHalls() ([]entities.Hall, error) {
	rows, err := r.db.Query(`
        SELECT id, cinema_id, name, rows, seats_per_row
        FROM halls ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var halls []entities.Hall
	for rows.Next() {
		var h entities.Hall
		if err := rows.Scan(&h.ID, &h.CinemaID, &h.Name, &h.Rows, &h.SeatsPerRow); err != nil {
			return nil, err
		}
		halls = append(halls, h)
	}
	return halls, rows.Err()
}

func (r *hallRepo) GenerateSeats(hallID uint, rows, seatsPerRow int) error {
	stmt, err := r.db.Prepare(`INSERT INTO seats (hall_id, row, number, reserved) VALUES ($1, $2, $3, false)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for row := 1; row <= rows; row++ {
		for num := 1; num <= seatsPerRow; num++ {
			_, err := stmt.Exec(hallID, row, num)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
