package repository

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"database/sql"
	"errors"
)

type seatsRepo struct {
	db *sql.DB
}

// NewRepo — конструктор
func NewSeatsRepo(db *sql.DB) ports.SeatsRepository {
	return &seatsRepo{db: db}
}

func (r *seatsRepo) GetSeats(sessionID int) ([]entities.Seat, error) {
	rows, err := r.db.Query(`
        SELECT s.id, s.row, s.col, b.id IS NOT NULL AS booked
        FROM seats s
        LEFT JOIN bookings b ON s.id = b.seat_id
        WHERE s.session_id = $1
        ORDER BY s.row, s.col`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []entities.Seat
	for rows.Next() {
		var s entities.Seat
		if err := rows.Scan(&s.ID, &s.HallID, &s.Row, &s.Number, &s.Reserved); err != nil {
			return nil, err
		}
		seats = append(seats, s)
	}
	return seats, nil
}

// BookSeat — бронирует место, если оно свободно
func (r *seatsRepo) BookSeat(userID, seatID int) error {
	tx, err := r.db.Begin() // ← r.db
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM bookings WHERE seat_id = $1)`, seatID).Scan(&exists) // ← r.db
	if err != nil {
		return err
	}
	if exists {
		return errors.New("seat already booked")
	}

	_, err = tx.Exec(`INSERT INTO bookings (user_id, seat_id) VALUES ($1, $2)`, userID, seatID) // ← r.db
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *seatsRepo) CreateSeat(hallID uint, row, number int) error {
	_, err := r.db.Exec(`
		INSERT INTO seats (hall_id, row, number, reserved) 
		VALUES ($1, $2, $3, false)`,
		hallID, row, number,
	)
	return err
}
